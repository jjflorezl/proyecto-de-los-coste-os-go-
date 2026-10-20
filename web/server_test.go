package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"costenos-narcolepsia/internal/scenario"
)

func testServer(t *testing.T) (*Server, http.Handler) {
	t.Helper()
	h, err := scenario.New()
	if err != nil {
		t.Fatal(err)
	}
	s, err := New(h)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	return s, s.Handler()
}

func request(handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}

func TestHTTPMandatoryScenario(t *testing.T) {
	s, handler := testServer(t)
	for i := 1; i <= 4; i++ {
		w := request(handler, "POST", "/api/attack", fmt.Sprintf(`{"patient_id":"P-%03d","location":"Lugar %d"}`, i, i))
		want := 200
		if i == 4 {
			want = 409
		}
		if w.Code != want {
			t.Fatalf("ataque %d: %d %s", i, w.Code, w.Body.String())
		}
	}
	if len(s.hospital.PatientsInHallway()) != 1 {
		t.Fatal("debe haber un paciente en pasillo")
	}
	for _, action := range []struct{ path, id string }{{"wake", "P-001"}, {"assign", "P-004"}} {
		w := request(handler, "POST", "/api/"+action.path, `{"patient_id":"`+action.id+`"}`)
		if w.Code != 200 {
			t.Fatal(w.Body.String())
		}
	}
	w := request(handler, "GET", "/api/state", "")
	var state struct {
		FreeBeds int            `json:"free_beds"`
		Episodes []any          `json:"episodes"`
		Severe   map[string]int `json:"severe"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if len(state.Episodes) != 4 || state.FreeBeds != 0 || state.Severe["P-005"] != 0 {
		t.Fatal("estado final incorrecto")
	}
}

func TestHTTPValidation(t *testing.T) {
	_, handler := testServer(t)
	for _, body := range []string{`null`, `{`, `{} {}`, `{"unknown":1}`} {
		t.Run(body, func(t *testing.T) {
			w := request(handler, "POST", "/api/simulation/start", body)
			if w.Code != 400 {
				t.Fatalf("esperaba 400 para %s; obtuvo %d", body, w.Code)
			}
		})
	}
	if w := request(handler, "GET", "/api/attack", ""); w.Code != 405 {
		t.Fatal("método incorrecto aceptado")
	}
	if w := request(handler, "POST", "/api/wake", `{"patient_id":"missing"}`); w.Code != 404 {
		t.Fatal("ID inexistente aceptado")
	}
	r := httptest.NewRequest("POST", "/api/simulation/start", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	if w.Code != 415 {
		t.Fatal("debe exigir JSON")
	}
}

func TestStaticAssets(t *testing.T) {
	_, handler := testServer(t)
	for _, path := range []string{"/", "/app.js", "/style.css"} {
		if w := request(handler, "GET", path, ""); w.Code != 200 || w.Body.Len() == 0 {
			t.Fatalf("recurso %s no disponible", path)
		}
	}
}

func TestSimulationWithConcurrentHTTPReaders(t *testing.T) {
	_, handler := testServer(t)
	if w := request(handler, "POST", "/api/simulation/start", `{}`); w.Code != 202 {
		t.Fatal(w.Body.String())
	}
	if w := request(handler, "POST", "/api/simulation/start", `{}`); w.Code != 409 {
		t.Fatal("se inició dos veces")
	}
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				if w := request(handler, "GET", "/api/state", ""); w.Code != 200 {
					t.Error("consulta fallida")
				}
			}
		}()
	}
	wg.Wait()
	if w := request(handler, "POST", "/api/simulation/stop", `{}`); w.Code != 200 {
		t.Fatal(w.Body.String())
	}
}
