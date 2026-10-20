package hospital

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"
)

type Hospital struct {
	mu            sync.Mutex
	name          string
	doctors       []*Doctor
	patients      []*Patient
	rooms         []*Room
	staff         []Attender
	history       []EpisodeRecord
	patientsByID  map[string]*Patient
	roomsByNumber map[int]*Room
	staffByID     map[string]Attender
	nextStaff     int
	nextEpisode   int
	now           func() time.Time
	zone          *time.Location
}

func NewHospital(name string) (*Hospital, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("nombre del hospital obligatorio")
	}
	return &Hospital{name: strings.TrimSpace(name), patientsByID: make(map[string]*Patient), roomsByNumber: make(map[int]*Room), staffByID: make(map[string]Attender), now: time.Now, zone: time.FixedZone("COT", -5*60*60)}, nil
}

func (h *Hospital) Name() string { return h.name }
func (h *Hospital) AdmitPatient(p *Patient) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if p == nil || p.hospital != h {
		return ErrNotAdmitted
	}
	if _, exists := h.patientsByID[p.ID()]; exists {
		return fmt.Errorf("ID de paciente duplicado: %s", p.ID())
	}
	h.patientsByID[p.ID()] = p
	h.patients = append(h.patients, p)
	return nil
}

func (h *Hospital) HireDoctor(d *Doctor) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if d == nil || d.hospital != h {
		return fmt.Errorf("médico de otro hospital o nulo")
	}
	if err := h.hireLocked(d); err != nil {
		return err
	}
	h.doctors = append(h.doctors, d)
	return nil
}
func (h *Hospital) HireOrderly(o *Orderly) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if o == nil || o.hospital != h {
		return fmt.Errorf("camillero de otro hospital o nulo")
	}
	return h.hireLocked(o)
}

// HireAttender is the extension point for another implementation, e.g. a nurse.
// Custom implementations must provide immutable identity and synchronize Attend.
func (h *Hospital) HireAttender(a Attender) error {
	if a == nil {
		return fmt.Errorf("personal nulo")
	}
	v := reflect.ValueOf(a)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return fmt.Errorf("personal nulo")
	}
	// Built-in types keep their ownership and doctor-list invariants.
	switch person := a.(type) {
	case *Doctor:
		return h.HireDoctor(person)
	case *Orderly:
		return h.HireOrderly(person)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.hireLocked(a)
}

func (h *Hospital) hireLocked(a Attender) error {
	if strings.TrimSpace(a.ID()) == "" || strings.TrimSpace(a.Name()) == "" {
		return fmt.Errorf("identidad del personal obligatoria")
	}
	if _, exists := h.staffByID[a.ID()]; exists {
		return fmt.Errorf("ID de personal duplicado: %s", a.ID())
	}
	h.staff = append(h.staff, a)
	h.staffByID[a.ID()] = a
	return nil
}

func (h *Hospital) RegisterRoom(r *Room) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if r == nil || r.hospital != h {
		return fmt.Errorf("habitación de otro hospital o nula")
	}
	if _, exists := h.roomsByNumber[r.number]; exists {
		return fmt.Errorf("habitación duplicada: %d", r.number)
	}
	h.roomsByNumber[r.number] = r
	h.rooms = append(h.rooms, r)
	return nil
}

func (h *Hospital) hasPatientLocked(p *Patient) bool {
	return p != nil && p.hospital == h && h.patientsByID[p.ID()] == p
}

// RegisterEpisode dispatches polymorphically. Attend performs the atomic operation.
// Staff work synchronously, so every completed turn is available for the next one.
func (h *Hospital) RegisterEpisode(p *Patient, location string) error {
	h.mu.Lock()
	if !h.hasPatientLocked(p) {
		h.mu.Unlock()
		return ErrNotAdmitted
	}
	if p.state != Awake {
		h.mu.Unlock()
		return ErrAlreadyAsleep
	}
	if strings.TrimSpace(location) == "" {
		h.mu.Unlock()
		return fmt.Errorf("ubicación obligatoria")
	}
	if len(h.staff) == 0 {
		h.mu.Unlock()
		return ErrNoStaff
	}
	a := h.staff[h.nextStaff%len(h.staff)]
	h.nextStaff++
	h.mu.Unlock() // Attend takes the same lock; holding it here would deadlock.
	_, err := a.Attend(p, location)
	return err
}

// Caller holds h.mu. Validation is repeated because dispatch releases the lock.
func (h *Hospital) attendLocked(p *Patient, location string, a Attender, d *Doctor) (EpisodeRecord, error) {
	if !h.hasPatientLocked(p) {
		return EpisodeRecord{}, ErrNotAdmitted
	}
	if p.state != Awake {
		return EpisodeRecord{}, ErrAlreadyAsleep
	}
	location = strings.TrimSpace(location)
	if location == "" {
		return EpisodeRecord{}, fmt.Errorf("ubicación obligatoria")
	}
	h.nextEpisode++
	p.state, p.currentLocation = AsleepInHallway, location
	record := EpisodeRecord{id: fmt.Sprintf("E-%04d", h.nextEpisode), date: h.now().In(h.zone), patient: p, attendingDoctor: d, attenderID: a.ID(), attenderName: a.Name(), location: location}
	p.pendingEpisode = record.id
	h.history = append(h.history, record)
	_, err := h.assignRoomLocked(p)
	return h.history[len(h.history)-1], err
}

func (h *Hospital) AssignRoom(p *Patient) (*Room, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.assignRoomLocked(p)
}

func (h *Hospital) assignRoomLocked(p *Patient) (*Room, error) {
	if !h.hasPatientLocked(p) {
		return nil, ErrNotAdmitted
	}
	if p.state != AsleepInHallway {
		return nil, ErrNotInHallway
	}
	for _, r := range h.rooms {
		if len(r.occupants) < r.capacity {
			if err := r.occupyLocked(p); err != nil {
				return nil, err
			}
			return r, nil
		}
	}
	return nil, ErrNoBed
}

func (h *Hospital) updateEpisodeRoomLocked(id string, room *Room) {
	for i := range h.history {
		if h.history[i].id == id {
			h.history[i].assignedRoom = room
		}
	}
	for _, d := range h.doctors {
		for i := range d.episodes {
			if d.episodes[i].id == id {
				d.episodes[i].assignedRoom = room
			}
		}
	}
}

func (h *Hospital) PatientsInHallway() []*Patient {
	h.mu.Lock()
	defer h.mu.Unlock()
	result := make([]*Patient, 0)
	for _, p := range h.patients {
		if p.state == AsleepInHallway {
			result = append(result, p)
		}
	}
	return result
}

func (h *Hospital) SevereReport() map[*Patient]int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.severeReportLocked()
}
func (h *Hospital) severeReportLocked() map[*Patient]int {
	result := make(map[*Patient]int)
	for _, p := range h.patients {
		if p.level == Severe {
			result[p] = 0
		}
	}
	y, m, day := h.now().In(h.zone).Date()
	for _, e := range h.history {
		ey, em, ed := e.date.In(h.zone).Date()
		if e.patient.level == Severe && y == ey && m == em && day == ed {
			result[e.patient]++
		}
	}
	return result
}

func (h *Hospital) Patients() []*Patient {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]*Patient{}, h.patients...)
}
func (h *Hospital) Doctors() []*Doctor {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]*Doctor{}, h.doctors...)
}
func (h *Hospital) Rooms() []*Room {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]*Room{}, h.rooms...)
}
func (h *Hospital) Staff() []Attender {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]Attender{}, h.staff...)
}
func (h *Hospital) History() []EpisodeRecord {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]EpisodeRecord{}, h.history...)
}
func (h *Hospital) PatientByID(id string) (*Patient, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	p, ok := h.patientsByID[id]
	if !ok {
		return nil, ErrNotAdmitted
	}
	return p, nil
}
