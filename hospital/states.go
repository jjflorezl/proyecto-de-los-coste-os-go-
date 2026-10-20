// Package hospital models the hospital independently of any user interface.
package hospital

import "errors"

type PatientState int

const (
	Awake PatientState = iota
	AsleepInHallway
	AsleepInBed
)

func (s PatientState) String() string {
	switch s {
	case Awake:
		return "Despierto"
	case AsleepInHallway:
		return "Dormido en pasillo"
	case AsleepInBed:
		return "Dormido en cama"
	default:
		return "Estado desconocido"
	}
}

type NarcolepsyLevel int

const (
	Mild NarcolepsyLevel = iota
	Moderate
	Severe
)

func (l NarcolepsyLevel) String() string {
	switch l {
	case Mild:
		return "Leve"
	case Moderate:
		return "Moderado"
	case Severe:
		return "Severo"
	default:
		return "Nivel desconocido"
	}
}

type RoomState int

const (
	Available RoomState = iota
	Occupied
)

func (s RoomState) String() string {
	switch s {
	case Available:
		return "Disponible"
	case Occupied:
		return "Ocupada"
	default:
		return "Estado desconocido"
	}
}

var (
	ErrNoBed         = errors.New("no hay cama disponible: el paciente permanece en el pasillo")
	ErrNotAdmitted   = errors.New("el paciente no está admitido en este hospital")
	ErrAlreadyAsleep = errors.New("el paciente ya está dormido")
	ErrAlreadyAwake  = errors.New("el paciente ya está despierto")
	ErrNotInHallway  = errors.New("el paciente debe estar dormido en el pasillo")
	ErrNoStaff       = errors.New("no hay personal contratado para atender")
)
