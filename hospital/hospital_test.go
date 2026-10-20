package hospital

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func fixture(t *testing.T, beds, patients int) (*Hospital, *Doctor, []*Patient) {
	t.Helper()
	h, err := NewHospital("Pruebas")
	if err != nil {
		t.Fatal(err)
	}
	d, err := NewDoctor(h, "D1", "Doctora", 35, "Neurología")
	if err != nil {
		t.Fatal(err)
	}
	if err := h.HireDoctor(d); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < beds; i++ {
		r, err := NewRoom(h, 101+i, 1)
		if err != nil {
			t.Fatal(err)
		}
		if err := h.RegisterRoom(r); err != nil {
			t.Fatal(err)
		}
	}
	ps := make([]*Patient, 0, patients)
	for i := 0; i < patients; i++ {
		p, err := NewPatient(h, fmt.Sprintf("P%d", i), "Paciente", 25, Severe)
		if err != nil {
			t.Fatal(err)
		}
		if err := h.AdmitPatient(p); err != nil {
			t.Fatal(err)
		}
		if err := d.DiagnosePatient(p); err != nil {
			t.Fatal(err)
		}
		ps = append(ps, p)
	}
	return h, d, ps
}

func TestAssignRoomAvailable(t *testing.T) {
	h, _, ps := fixture(t, 1, 1)
	if err := ps[0].SufferSleepAttack("Cafetería"); err != nil {
		t.Fatal(err)
	}
	if ps[0].State() != AsleepInBed || ps[0].AssignedRoom().Number() != 101 {
		t.Fatal("asignación incorrecta")
	}
	if h.Rooms()[0].State() != Occupied || h.Rooms()[0].IsAvailable() {
		t.Fatal("habitación debe estar llena")
	}
	if len(h.History()) != 1 || h.History()[0].AssignedRoom() == nil {
		t.Fatal("episodio no registrado")
	}
}

func TestAssignRoomNoAvailability(t *testing.T) {
	h, d, ps := fixture(t, 1, 2)
	if err := ps[0].SufferSleepAttack("Jardín"); err != nil {
		t.Fatal(err)
	}
	if err := ps[1].SufferSleepAttack("Pasillo"); !errors.Is(err, ErrNoBed) {
		t.Fatalf("error: %v", err)
	}
	r, err := h.AssignRoom(ps[1])
	if r != nil || !errors.Is(err, ErrNoBed) || ps[1].State() != AsleepInHallway {
		t.Fatal("debe permanecer en pasillo sin habitación inventada")
	}
	if len(h.History()) != 2 || len(d.MyEpisodes()) != 2 {
		t.Fatal("el episodio sin cama también debe registrarse")
	}
}

func TestHallwayQueryAndReassignment(t *testing.T) {
	h, d, ps := fixture(t, 1, 2)
	if err := ps[0].SufferSleepAttack("Jardín"); err != nil {
		t.Fatal(err)
	}
	if err := ps[1].SufferSleepAttack("Pasillo"); !errors.Is(err, ErrNoBed) {
		t.Fatal(err)
	}
	if got := h.PatientsInHallway(); len(got) != 1 || got[0] != ps[1] {
		t.Fatal("consulta incorrecta")
	}
	if err := ps[0].WakeUp(); err != nil {
		t.Fatal(err)
	}
	if !h.Rooms()[0].IsAvailable() || ps[0].AssignedRoom() != nil {
		t.Fatal("despertar debe liberar")
	}
	if _, err := h.AssignRoom(ps[1]); err != nil {
		t.Fatal(err)
	}
	if len(h.PatientsInHallway()) != 0 || len(h.History()) != 2 {
		t.Fatal("reasignar no crea episodios")
	}
	if h.History()[1].AssignedRoom() == nil || d.MyEpisodes()[1].AssignedRoom() == nil {
		t.Fatal("ambos historiales deben reflejar la cama asignada después")
	}
	// Historical bed remains even after its previous occupant woke up.
	if h.History()[0].AssignedRoom() == nil {
		t.Fatal("se perdió el dato histórico")
	}
}

func TestSevereReportOnlyTodayIncludingZero(t *testing.T) {
	h, _, ps := fixture(t, 1, 2)
	now := time.Date(2026, 10, 19, 4, 59, 0, 0, time.UTC) // Oct 18, 23:59 in Colombia.
	h.now = func() time.Time { return now }
	if err := ps[0].SufferSleepAttack("Ayer"); err != nil {
		t.Fatal(err)
	}
	if err := ps[0].WakeUp(); err != nil {
		t.Fatal(err)
	}
	now = now.Add(2 * time.Minute)
	if err := ps[0].SufferSleepAttack("Hoy"); err != nil {
		t.Fatal(err)
	}
	got := h.SevereReport()
	if got[ps[0]] != 1 || got[ps[1]] != 0 || len(got) != 2 {
		t.Fatalf("reporte: %v", got)
	}
}

func TestPolymorphicDispatch(t *testing.T) {
	h, d, ps := fixture(t, 2, 2)
	o, err := NewOrderly(h, "C1", "Camillero", 30)
	if err != nil {
		t.Fatal(err)
	}
	if err := h.HireOrderly(o); err != nil {
		t.Fatal(err)
	}
	for _, p := range ps {
		if err := p.SufferSleepAttack("Recepción"); err != nil {
			t.Fatal(err)
		}
	}
	history := h.History()
	if history[0].AttenderID() != d.ID() || history[1].AttenderID() != o.ID() {
		t.Fatal("no se despacharon los dos tipos")
	}
	if history[1].Doctor() != nil || len(d.MyEpisodes()) != 1 {
		t.Fatal("camillero no debe suplantar un médico")
	}
}

func TestCapacityAndRelease(t *testing.T) {
	h, _, ps := fixture(t, 0, 3)
	r, err := NewRoom(h, 202, 2)
	if err != nil {
		t.Fatal(err)
	}
	if err := h.RegisterRoom(r); err != nil {
		t.Fatal(err)
	}
	if err := ps[0].SufferSleepAttack("A"); err != nil {
		t.Fatal(err)
	}
	if !r.IsAvailable() || r.State() != Occupied {
		t.Fatal("queda una cama en una habitación ocupada")
	}
	if err := ps[1].SufferSleepAttack("B"); err != nil {
		t.Fatal(err)
	}
	if err := ps[2].SufferSleepAttack("C"); !errors.Is(err, ErrNoBed) {
		t.Fatal(err)
	}
	if err := r.Release(ps[0]); err != nil {
		t.Fatal(err)
	}
	if ps[0].State() != AsleepInHallway || !r.IsAvailable() {
		t.Fatal("liberación inconsistente")
	}
	if err := r.Release(ps[0]); err == nil {
		t.Fatal("doble liberación debe fallar")
	}
	if err := r.Occupy(ps[2]); err != nil {
		t.Fatal(err)
	}
}

func TestValidationAndCopies(t *testing.T) {
	h, d, ps := fixture(t, 1, 1)
	if err := h.AdmitPatient(ps[0]); err == nil {
		t.Fatal("duplicado admitido")
	}
	if err := h.HireDoctor(d); err == nil {
		t.Fatal("médico duplicado")
	}
	if err := h.RegisterEpisode(nil, "A"); !errors.Is(err, ErrNotAdmitted) {
		t.Fatal(err)
	}
	if err := ps[0].SufferSleepAttack(" "); err == nil {
		t.Fatal("ubicación vacía")
	}
	if err := ps[0].WakeUp(); !errors.Is(err, ErrAlreadyAwake) {
		t.Fatal(err)
	}
	if err := ps[0].SufferSleepAttack("A"); err != nil {
		t.Fatal(err)
	}
	if err := ps[0].SufferSleepAttack("B"); !errors.Is(err, ErrAlreadyAsleep) {
		t.Fatal(err)
	}
	copyPatients := h.Patients()
	copyPatients[0] = nil
	copyEpisodes := d.MyEpisodes()
	copyEpisodes[0] = EpisodeRecord{}
	if h.Patients()[0] == nil || d.MyEpisodes()[0].ID() == "" {
		t.Fatal("se expusieron slices internos")
	}
	other, _, _ := fixture(t, 0, 0)
	if err := other.AdmitPatient(ps[0]); err == nil {
		t.Fatal("se admitió paciente de otro hospital")
	}
	if _, err := NewPatient(h, "", "A", 20, Mild); err == nil {
		t.Fatal("identidad inválida")
	}
	if _, err := NewRoom(h, 1, 0); err == nil {
		t.Fatal("capacidad inválida")
	}
}

func TestConcurrentLastBedAndReaders(t *testing.T) {
	h, _, ps := fixture(t, 1, 20)
	var wg sync.WaitGroup
	for _, p := range ps {
		wg.Add(1)
		go func(p *Patient) {
			defer wg.Done()
			if err := p.SufferSleepAttack("Pasillo"); err != nil && !errors.Is(err, ErrNoBed) {
				t.Error(err)
			}
			_ = h.Snapshot()
			_ = h.SevereReport()
			_ = p.State()
			_ = h.Doctors()[0].MyEpisodes()
		}(p)
	}
	wg.Wait()
	if len(h.Rooms()[0].Occupants()) != 1 || len(h.PatientsInHallway()) != 19 || len(h.History()) != 20 {
		t.Fatal("se violó capacidad o se perdió un episodio")
	}
}

func TestDoctorRescueAndDiagnosis(t *testing.T) {
	h, d, ps := fixture(t, 0, 1)
	if err := ps[0].SufferSleepAttack("Pasillo"); !errors.Is(err, ErrNoBed) {
		t.Fatal(err)
	}
	r, err := NewRoom(h, 101, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := h.RegisterRoom(r); err != nil {
		t.Fatal(err)
	}
	if err := d.AttendSleepEmergency(ps[0]); err != nil {
		t.Fatal(err)
	}
	if len(d.MyEpisodes()) != 1 || ps[0].State() != AsleepInBed {
		t.Fatal("rescate debe conservar un solo episodio")
	}
	if ps[0].AssignedDoctor() != d || len(d.Patients()) != 1 {
		t.Fatal("relación médico-paciente incorrecta")
	}
}
