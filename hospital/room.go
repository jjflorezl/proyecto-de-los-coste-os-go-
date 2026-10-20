package hospital

import "fmt"

type Room struct {
	hospital  *Hospital
	number    int
	capacity  int
	state     RoomState
	occupants []*Patient
}

func NewRoom(h *Hospital, number, capacity int) (*Room, error) {
	if h == nil || number <= 0 || capacity <= 0 {
		return nil, fmt.Errorf("hospital, número y capacidad deben ser válidos")
	}
	return &Room{hospital: h, number: number, capacity: capacity, state: Available}, nil
}

func (r *Room) Number() int   { return r.number }
func (r *Room) Capacity() int { return r.capacity }
func (r *Room) State() RoomState {
	r.hospital.mu.Lock()
	defer r.hospital.mu.Unlock()
	return r.state
}
func (r *Room) IsAvailable() bool {
	r.hospital.mu.Lock()
	defer r.hospital.mu.Unlock()
	return len(r.occupants) < r.capacity
}
func (r *Room) Occupants() []*Patient {
	r.hospital.mu.Lock()
	defer r.hospital.mu.Unlock()
	return append([]*Patient{}, r.occupants...)
}

func (r *Room) Occupy(p *Patient) error {
	r.hospital.mu.Lock()
	defer r.hospital.mu.Unlock()
	return r.occupyLocked(p)
}

func (r *Room) occupyLocked(p *Patient) error {
	h := r.hospital
	if h.roomsByNumber[r.number] != r {
		return fmt.Errorf("habitación no registrada")
	}
	if !h.hasPatientLocked(p) {
		return ErrNotAdmitted
	}
	if p.state != AsleepInHallway || p.room != nil {
		return ErrNotInHallway
	}
	if len(r.occupants) >= r.capacity {
		return ErrNoBed
	}
	r.occupants = append(r.occupants, p)
	r.state = Occupied // Occupied means at least one occupant; IsAvailable checks capacity.
	p.room, p.state, p.currentLocation = r, AsleepInBed, fmt.Sprintf("Habitación %d", r.number)
	h.updateEpisodeRoomLocked(p.pendingEpisode, r)
	return nil
}

func (r *Room) Release(p *Patient) error {
	r.hospital.mu.Lock()
	defer r.hospital.mu.Unlock()
	return r.releaseLocked(p)
}

func (r *Room) releaseLocked(p *Patient) error {
	if !r.hospital.hasPatientLocked(p) {
		return ErrNotAdmitted
	}
	for i, occupant := range r.occupants {
		if occupant != p {
			continue
		}
		r.occupants = append(r.occupants[:i], r.occupants[i+1:]...)
		if len(r.occupants) == 0 {
			r.state = Available
		}
		p.room = nil
		if p.state == AsleepInBed {
			p.state = AsleepInHallway
		}
		p.currentLocation = fmt.Sprintf("Pasillo frente a habitación %d", r.number)
		return nil
	}
	return fmt.Errorf("el paciente no ocupa la habitación %d", r.number)
}
