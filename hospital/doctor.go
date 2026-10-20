package hospital

import (
	"fmt"
	"strings"
)

type Attender interface {
	ID() string
	Name() string
	Attend(p *Patient, location string) (EpisodeRecord, error)
}

type Doctor struct {
	Person
	hospital  *Hospital
	specialty string
	patients  []*Patient
	episodes  []EpisodeRecord
}

var _ Attender = (*Doctor)(nil)

func NewDoctor(h *Hospital, id, name string, age int, specialty string) (*Doctor, error) {
	if h == nil || strings.TrimSpace(specialty) == "" {
		return nil, fmt.Errorf("hospital y especialidad obligatorios")
	}
	person, err := newPerson(id, name, age)
	if err != nil {
		return nil, err
	}
	return &Doctor{Person: person, hospital: h, specialty: strings.TrimSpace(specialty)}, nil
}

func (d *Doctor) Specialty() string { return d.specialty }
func (d *Doctor) Patients() []*Patient {
	d.hospital.mu.Lock()
	defer d.hospital.mu.Unlock()
	return append([]*Patient{}, d.patients...)
}
func (d *Doctor) MyEpisodes() []EpisodeRecord {
	d.hospital.mu.Lock()
	defer d.hospital.mu.Unlock()
	return append([]EpisodeRecord{}, d.episodes...)
}

func (d *Doctor) DiagnosePatient(p *Patient) error {
	h := d.hospital
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.hasPatientLocked(p) {
		return ErrNotAdmitted
	}
	if h.staffByID[d.ID()] != d {
		return fmt.Errorf("médico no contratado")
	}
	if p.assignedDoctor == d {
		return nil
	}
	if old := p.assignedDoctor; old != nil {
		for i, item := range old.patients {
			if item == p {
				old.patients = append(old.patients[:i], old.patients[i+1:]...)
				break
			}
		}
	}
	p.assignedDoctor = d
	d.patients = append(d.patients, p)
	return nil
}

func (d *Doctor) Attend(p *Patient, location string) (EpisodeRecord, error) {
	h := d.hospital
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.staffByID[d.ID()] != d {
		return EpisodeRecord{}, fmt.Errorf("médico no contratado")
	}
	record, err := h.attendLocked(p, location, d, d)
	if record.id != "" {
		d.episodes = append(d.episodes, record)
	}
	return record, err
}

// AttendSleepEmergency rescues an existing hallway episode without counting a new attack.
func (d *Doctor) AttendSleepEmergency(p *Patient) error {
	h := d.hospital
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.hasPatientLocked(p) {
		return ErrNotAdmitted
	}
	if h.staffByID[d.ID()] != d {
		return fmt.Errorf("médico no contratado")
	}
	if p.state != AsleepInHallway {
		return ErrNotInHallway
	}
	_, err := h.assignRoomLocked(p)
	for _, record := range h.history {
		if record.id != p.pendingEpisode {
			continue
		}
		for _, existing := range d.episodes {
			if existing.id == record.id {
				return err
			}
		}
		// Preserve the original dispatcher in the hospital history; this doctor also intervened.
		record.attendingDoctor = d
		d.episodes = append(d.episodes, record)
		break
	}
	return err
}
