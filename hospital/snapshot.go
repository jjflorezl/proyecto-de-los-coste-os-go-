package hospital

// Snapshot is a consistent, detached read model, useful for any consumer.
// It has no HTTP, JSON, HTML, or presentation dependencies.
type Snapshot struct {
	Name         string
	Patients     []PatientSnapshot
	Rooms        []RoomSnapshot
	Staff        []StaffSnapshot
	History      []EpisodeRecord
	SevereCounts map[string]int
}
type PatientSnapshot struct {
	ID, Name                       string
	Age                            int
	Level                          NarcolepsyLevel
	State                          PatientState
	Location, DoctorID, DoctorName string
	RoomNumber                     int
}
type RoomSnapshot struct {
	Number, Capacity int
	State            RoomState
	OccupantIDs      []string
}
type StaffSnapshot struct {
	ID, Name string
}

func (h *Hospital) Snapshot() Snapshot {
	h.mu.Lock()
	defer h.mu.Unlock()
	s := Snapshot{Name: h.name, Patients: []PatientSnapshot{}, Rooms: []RoomSnapshot{}, Staff: []StaffSnapshot{}, History: append([]EpisodeRecord{}, h.history...), SevereCounts: map[string]int{}}
	for _, p := range h.patients {
		item := PatientSnapshot{ID: p.ID(), Name: p.Name(), Age: p.Age(), Level: p.level, State: p.state, Location: p.currentLocation}
		if p.assignedDoctor != nil {
			item.DoctorID, item.DoctorName = p.assignedDoctor.ID(), p.assignedDoctor.Name()
		}
		if p.room != nil {
			item.RoomNumber = p.room.number
		}
		s.Patients = append(s.Patients, item)
	}
	for _, r := range h.rooms {
		item := RoomSnapshot{Number: r.number, Capacity: r.capacity, State: r.state, OccupantIDs: []string{}}
		for _, p := range r.occupants {
			item.OccupantIDs = append(item.OccupantIDs, p.ID())
		}
		s.Rooms = append(s.Rooms, item)
	}
	for _, a := range h.staff {
		s.Staff = append(s.Staff, StaffSnapshot{ID: a.ID(), Name: a.Name()})
	}
	for p, count := range h.severeReportLocked() {
		s.SevereCounts[p.ID()] = count
	}
	return s
}
