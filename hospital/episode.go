package hospital

import (
	"fmt"
	"time"
)

// EpisodeRecord is a value snapshot. A nil doctor means an orderly attended.
// assignedDoctor on Patient is a separate relationship from the episode doctor.
type EpisodeRecord struct {
	id              string
	date            time.Time
	patient         *Patient
	attendingDoctor *Doctor
	attenderID      string
	attenderName    string
	location        string
	assignedRoom    *Room
}

func (e EpisodeRecord) ID() string           { return e.id }
func (e EpisodeRecord) Date() time.Time      { return e.date }
func (e EpisodeRecord) Patient() *Patient    { return e.patient }
func (e EpisodeRecord) Doctor() *Doctor      { return e.attendingDoctor }
func (e EpisodeRecord) AttenderID() string   { return e.attenderID }
func (e EpisodeRecord) AttenderName() string { return e.attenderName }
func (e EpisodeRecord) Location() string     { return e.location }
func (e EpisodeRecord) AssignedRoom() *Room  { return e.assignedRoom }

func (e EpisodeRecord) Summary() string {
	if e.patient == nil {
		return "Registro vacío"
	}
	bed := "sin cama"
	if e.assignedRoom != nil {
		bed = fmt.Sprintf("habitación %d", e.assignedRoom.Number())
	}
	return fmt.Sprintf("%s | %s | %s | %s -> %s | %s", e.id, e.date.Format("2006-01-02 15:04:05"), e.patient.ID(), e.location, bed, e.attenderName)
}
