package hospital

import "fmt"

// Orderly satisfies Attender but deliberately has no DiagnosePatient method.
type Orderly struct {
	Person
	hospital *Hospital
}

var _ Attender = (*Orderly)(nil)

func NewOrderly(h *Hospital, id, name string, age int) (*Orderly, error) {
	if h == nil {
		return nil, fmt.Errorf("hospital obligatorio")
	}
	person, err := newPerson(id, name, age)
	if err != nil {
		return nil, err
	}
	return &Orderly{Person: person, hospital: h}, nil
}

func (o *Orderly) Attend(p *Patient, location string) (EpisodeRecord, error) {
	h := o.hospital
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.staffByID[o.ID()] != o {
		return EpisodeRecord{}, fmt.Errorf("camillero no contratado")
	}
	if !h.hasPatientLocked(p) {
		return EpisodeRecord{}, ErrNotAdmitted
	}
	// A transfer is not a second sleep attack. Keep the original record and dispatcher.
	if p.state == AsleepInHallway {
		_, err := h.assignRoomLocked(p)
		for _, record := range h.history {
			if record.id == p.pendingEpisode {
				return record, err
			}
		}
		return EpisodeRecord{}, fmt.Errorf("paciente sin episodio pendiente")
	}
	return h.attendLocked(p, location, o, nil)
}
