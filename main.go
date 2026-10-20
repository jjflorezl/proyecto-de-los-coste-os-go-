package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"costenos-narcolepsia/hospital"
	"costenos-narcolepsia/internal/scenario"
	"costenos-narcolepsia/simulation"
)

func main() {
	simulate := flag.Bool("simulate", false, "ejecutar el bonus concurrente después del escenario")
	duration := flag.Duration("duration", 8*time.Second, "duración de la simulación")
	seed := flag.Int64("seed", 42, "semilla de los intervalos aleatorios")
	flag.Parse()
	if err := run(*simulate, *duration, *seed); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run(simulate bool, duration time.Duration, seed int64) error {
	if simulate && duration <= 0 {
		return fmt.Errorf("la duración debe ser positiva")
	}
	h, err := scenario.New()
	if err != nil {
		return err
	}
	fmt.Println(h.Name())
	fmt.Println("Autor: Juan José Flórez López")
	fmt.Println("\n== Escenario obligatorio: 2 médicos, 1 camillero, 3 camas y 5 pacientes ==")
	patients := h.Patients()
	locations := []string{"Cafetería", "Jardín", "Fila de radiología", "Pasillo 2, segundo piso"}
	for i, location := range locations {
		if err := patients[i].SufferSleepAttack(location); err != nil {
			if !errors.Is(err, hospital.ErrNoBed) {
				return err
			}
			fmt.Printf("%s: %v\n", patients[i].ID(), err)
		} else {
			fmt.Println(h.History()[i].Summary())
		}
	}
	fmt.Printf("Pacientes en pasillo antes de liberar una cama: %d\n", len(h.PatientsInHallway()))
	freed := patients[0].AssignedRoom()
	if err := patients[0].WakeUp(); err != nil {
		return err
	}
	fmt.Printf("%s despierta y libera la habitación %d.\n", patients[0].ID(), freed.Number())
	room, err := h.AssignRoom(patients[3])
	if err != nil {
		return err
	}
	fmt.Printf("%s pasa del pasillo a la habitación %d.\n", patients[3].ID(), room.Number())

	fmt.Println("\n== Consulta 1: pacientes dormidos en pasillo (estado final) ==")
	hallway := h.PatientsInHallway()
	if len(hallway) == 0 {
		fmt.Println("Ninguno: el paciente pendiente ya recibió una cama.")
	}
	for _, p := range hallway {
		fmt.Printf("%s | %s | %s\n", p.ID(), p.Name(), p.CurrentLocation())
	}
	fmt.Println("\n== Consulta 2: historial por médico ==")
	for _, d := range h.Doctors() {
		fmt.Println("Médico:", d.Name())
		for _, record := range d.MyEpisodes() {
			fmt.Println(" ", record.Summary())
		}
	}
	fmt.Println("\n== Consulta 3: disponibilidad y asignación de cama ==")
	fmt.Println("La cuarta asignación falló; tras despertar P-001, AssignRoom(P-004) devolvió la habitación", room.Number())
	for _, r := range h.Rooms() {
		fmt.Printf("Habitación %d: %s | camas libres: %d\n", r.Number(), r.State(), r.Capacity()-len(r.Occupants()))
	}
	fmt.Println("\n== Consulta 4: reporte de pacientes severos, hoy (Colombia) ==")
	report := h.SevereReport()
	for _, p := range patients {
		if count, ok := report[p]; ok {
			fmt.Printf("%s | %-20s | %d episodios hoy\n", p.ID(), p.Name(), count)
		}
	}
	fmt.Printf("\nTotal: %d episodios; el camillero atendió %s.\n", len(h.History()), h.History()[2].Patient().ID())
	if simulate {
		fmt.Println("\n== Bonus: una goroutine por paciente ==")
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		ctx, cancel := context.WithTimeout(ctx, duration)
		defer cancel()
		stats, err := simulation.Run(ctx, h, simulation.Config{MinInterval: 100 * time.Millisecond, MaxInterval: 450 * time.Millisecond, Seed: seed})
		if err != nil {
			return err
		}
		fmt.Printf("Simulación finalizada: %d ataques, %d intentos sin cama, %d despertares.\n", stats.Attacks, stats.BedShortages, stats.WakeUps)
		fmt.Printf("Historial final: %d episodios.\n", len(h.History()))
	}
	return nil
}
