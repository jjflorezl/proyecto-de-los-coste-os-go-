package hospital

import "fmt"

type Patient struct {
	Person
	hospital        *Hospital
	level           NarcolepsyLevel
	state           PatientState
	currentLocation string
	assignedDoctor  *Doctor
	room            *Room
	pendingEpisode  string
}

// Entities belong to one hospital from construction; its mutex protects them.
func NewPatient(h *Hospital, id, name string, age int, level NarcolepsyLevel) (*Patient, error) {
	if h == nil || level < Mild || level > Severe {
		return nil, fmt.Errorf("hospital o nivel inválido")
	}
	person, err := newPerson(id, name, age)
	if err != nil {
		return nil, err
	}
	return &Patient{Person: person, hospital: h, level: level, state: Awake, currentLocation: "Recepción"}, nil
}

// SufferSleepAttack records the attack even if no bed is available.
func (p *Patient) SufferSleepAttack(location string) error {
	return p.hospital.RegisterEpisode(p, location)
}

// WakeUp releases the bed in the same transaction as the state change.
func (p *Patient) WakeUp() error {
	h := p.hospital
	h.mu.Lock()
	defer h.mu.Unlock()
	if !h.hasPatientLocked(p) {
		return ErrNotAdmitted
	}
	if p.state == Awake {
		return ErrAlreadyAwake
	}
	if p.room != nil {
		if err := p.room.releaseLocked(p); err != nil {
			return err
		}
	}
	p.state, p.currentLocation, p.pendingEpisode = Awake, "Recepción", ""
	return nil
}

func (p *Patient) State() PatientState {
	p.hospital.mu.Lock()
	defer p.hospital.mu.Unlock()
	return p.state
}
func (p *Patient) Level() NarcolepsyLevel { return p.level }
func (p *Patient) CurrentLocation() string {
	p.hospital.mu.Lock()
	defer p.hospital.mu.Unlock()
	return p.currentLocation
}
func (p *Patient) AssignedDoctor() *Doctor {
	p.hospital.mu.Lock()
	defer p.hospital.mu.Unlock()
	return p.assignedDoctor
}
func (p *Patient) AssignedRoom() *Room {
	p.hospital.mu.Lock()
	defer p.hospital.mu.Unlock()
	return p.room
}
