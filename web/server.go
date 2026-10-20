// Package web adapts the hospital model to HTTP. All display code lives here.
package web

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"strings"
	"sync"
	"time"

	"costenos-narcolepsia/hospital"
	"costenos-narcolepsia/simulation"
)

//go:embed static/*
var assets embed.FS

type Server struct {
	hospital        *hospital.Hospital
	mu              sync.Mutex
	running, closed bool
	cancel          context.CancelFunc
	done            chan struct{}
	lastError       string
	stats           simulation.Stats
}

func New(h *hospital.Hospital) (*Server, error) {
	if h == nil {
		return nil, fmt.Errorf("hospital obligatorio")
	}
	return &Server{hospital: h}, nil
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/state", s.state)
	mux.HandleFunc("/api/attack", s.action("attack"))
	mux.HandleFunc("/api/wake", s.action("wake"))
	mux.HandleFunc("/api/assign", s.action("assign"))
	mux.HandleFunc("/api/simulation/start", s.start)
	mux.HandleFunc("/api/simulation/stop", s.stop)
	// The embedded directory exists at build time; no disk files are needed at runtime.
	mux.Handle("/", http.FileServer(http.FS(mustStatic())))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self'; script-src 'self'; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		mux.ServeHTTP(w, r)
	})
}

func mustStatic() fs.FS {
	// fs.Sub only rejects invalid path syntax; "static" is a constant valid path.
	static, _ := fs.Sub(assets, "static")
	return static
}

func send(w http.ResponseWriter, code int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(value)
}

func method(w http.ResponseWriter, r *http.Request, expected string) bool {
	if r.Method == expected {
		return true
	}
	w.Header().Set("Allow", expected)
	send(w, http.StatusMethodNotAllowed, map[string]string{"error": "método no permitido"})
	return false
}

type command struct {
	PatientID string `json:"patient_id"`
	Location  string `json:"location"`
}

func readCommand(w http.ResponseWriter, r *http.Request) (command, bool) {
	var input command
	if !method(w, r, http.MethodPost) {
		return input, false
	}
	contentType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || contentType != "application/json" {
		send(w, http.StatusUnsupportedMediaType, map[string]string{"error": "se requiere application/json"})
		return input, false
	}
	if origin := r.Header.Get("Origin"); origin != "" && origin != "http://"+r.Host && origin != "https://"+r.Host {
		send(w, http.StatusForbidden, map[string]string{"error": "origen no permitido"})
		return input, false
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	var decoded *command
	if err := decoder.Decode(&decoded); err != nil || decoded == nil {
		send(w, http.StatusBadRequest, map[string]string{"error": "JSON inválido"})
		return input, false
	}
	input = *decoded
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		send(w, http.StatusBadRequest, map[string]string{"error": "se esperaba un único objeto JSON"})
		return input, false
	}
	return input, true
}

func (s *Server) action(action string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		input, ok := readCommand(w, r)
		if !ok {
			return
		}
		p, err := s.hospital.PatientByID(input.PatientID)
		if err != nil {
			send(w, 404, map[string]string{"error": err.Error()})
			return
		}
		switch action {
		case "attack":
			err = p.SufferSleepAttack(input.Location)
		case "wake":
			err = p.WakeUp()
		case "assign":
			_, err = s.hospital.AssignRoom(p)
		}
		if err != nil {
			code := http.StatusConflict
			if strings.TrimSpace(input.Location) == "" && action == "attack" {
				code = http.StatusBadRequest
			}
			send(w, code, map[string]any{"error": err.Error(), "episode_recorded": errors.Is(err, hospital.ErrNoBed) && action == "attack"})
			return
		}
		send(w, http.StatusOK, map[string]string{"message": "Operación realizada"})
	}
}

func (s *Server) state(w http.ResponseWriter, r *http.Request) {
	if !method(w, r, http.MethodGet) {
		return
	}
	snapshot := s.hospital.Snapshot()
	patients, rooms, staff, episodes := []any{}, []any{}, []any{}, []any{}
	for _, p := range snapshot.Patients {
		patients = append(patients, map[string]any{"id": p.ID, "name": p.Name, "age": p.Age, "level": p.Level.String(), "state": p.State.String(), "state_code": int(p.State), "location": p.Location, "doctor": p.DoctorName, "room": p.RoomNumber})
	}
	free := 0
	for _, room := range snapshot.Rooms {
		available := room.Capacity - len(room.OccupantIDs)
		free += available
		rooms = append(rooms, map[string]any{"number": room.Number, "capacity": room.Capacity, "free": available, "state": room.State.String(), "occupants": room.OccupantIDs})
	}
	for _, a := range snapshot.Staff {
		staff = append(staff, map[string]string{"id": a.ID, "name": a.Name})
	}
	for _, e := range snapshot.History {
		room := 0
		if e.AssignedRoom() != nil {
			room = e.AssignedRoom().Number()
		}
		episodes = append(episodes, map[string]any{"id": e.ID(), "date": e.Date().Format(time.RFC3339), "patient": e.Patient().ID(), "patient_name": e.Patient().Name(), "attender": e.AttenderName(), "attender_id": e.AttenderID(), "location": e.Location(), "room": room})
	}
	s.mu.Lock()
	running, lastError, stats := s.running, s.lastError, s.stats
	s.mu.Unlock()
	send(w, http.StatusOK, map[string]any{"name": snapshot.Name, "patients": patients, "rooms": rooms, "staff": staff, "episodes": episodes, "severe": snapshot.SevereCounts, "free_beds": free, "simulation_running": running, "simulation_error": lastError, "simulation_stats": stats})
}

func (s *Server) start(w http.ResponseWriter, r *http.Request) {
	if _, ok := readCommand(w, r); !ok {
		return
	}
	s.mu.Lock()
	if s.running || s.closed {
		s.mu.Unlock()
		send(w, 409, map[string]string{"error": "simulación activa o servidor cerrado"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	done := make(chan struct{})
	s.cancel, s.done, s.running, s.lastError = cancel, done, true, ""
	s.mu.Unlock()
	go func() {
		stats, err := simulation.Run(ctx, s.hospital, simulation.Config{MinInterval: 900 * time.Millisecond, MaxInterval: 2500 * time.Millisecond, Seed: time.Now().UnixNano()})
		cancel()
		s.mu.Lock()
		s.running, s.stats = false, stats
		if err != nil {
			s.lastError = err.Error()
		}
		s.mu.Unlock()
		close(done)
	}()
	send(w, http.StatusAccepted, map[string]string{"message": "Simulación iniciada durante 30 segundos"})
}

func (s *Server) stopSimulation() {
	s.mu.Lock()
	cancel, done := s.cancel, s.done
	s.mu.Unlock()
	if cancel != nil {
		cancel()
		<-done
	}
}

func (s *Server) stop(w http.ResponseWriter, r *http.Request) {
	if _, ok := readCommand(w, r); !ok {
		return
	}
	s.stopSimulation()
	send(w, http.StatusOK, map[string]string{"message": "Simulación detenida"})
}

func (s *Server) Close() {
	s.mu.Lock()
	s.closed = true
	s.mu.Unlock()
	s.stopSimulation()
}
