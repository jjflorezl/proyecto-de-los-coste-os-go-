// Package scenario builds the same starting data for the console and web adapter.
package scenario

import "costenos-narcolepsia/hospital"

func New() (*hospital.Hospital, error) {
	h, err := hospital.NewHospital("Hospital de los Costeños con Narcolepsia")
	if err != nil {
		return nil, err
	}
	for _, spec := range []struct {
		id, name, specialty string
		age                 int
	}{
		{"D-001", "Karen Ospina", "Neurología", 38},
		{"D-002", "Rafael Pacheco", "Medicina del sueño", 45},
	} {
		d, err := hospital.NewDoctor(h, spec.id, spec.name, spec.age, spec.specialty)
		if err != nil {
			return nil, err
		}
		if err := h.HireDoctor(d); err != nil {
			return nil, err
		}
	}
	o, err := hospital.NewOrderly(h, "C-001", "Luis Mercado", 29)
	if err != nil {
		return nil, err
	}
	if err := h.HireOrderly(o); err != nil {
		return nil, err
	}
	for _, number := range []int{101, 102, 103} {
		r, err := hospital.NewRoom(h, number, 1)
		if err != nil {
			return nil, err
		}
		if err := h.RegisterRoom(r); err != nil {
			return nil, err
		}
	}
	for i, spec := range []struct {
		id, name string
		age      int
		level    hospital.NarcolepsyLevel
	}{
		{"P-001", "Carlos Vives", 34, hospital.Moderate},
		{"P-002", "Yeimy Padilla", 27, hospital.Severe},
		{"P-003", "Andrés Salcedo", 42, hospital.Mild},
		{"P-004", "Wilfrido Berrio", 51, hospital.Severe},
		{"P-005", "María Camila Torres", 23, hospital.Severe},
	} {
		p, err := hospital.NewPatient(h, spec.id, spec.name, spec.age, spec.level)
		if err != nil {
			return nil, err
		}
		if err := h.AdmitPatient(p); err != nil {
			return nil, err
		}
		if err := h.Doctors()[i%2].DiagnosePatient(p); err != nil {
			return nil, err
		}
	}
	return h, nil
}
