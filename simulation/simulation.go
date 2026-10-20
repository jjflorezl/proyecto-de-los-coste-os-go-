// Package simulation drives the same domain model with one goroutine per patient.
package simulation

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"costenos-narcolepsia/hospital"
)

type Config struct {
	MinInterval, MaxInterval time.Duration
	Seed                     int64
}
type Stats struct{ Attacks, BedShortages, WakeUps int }
type result struct {
	stats Stats
	err   error
}

func Run(ctx context.Context, h *hospital.Hospital, config Config) (Stats, error) {
	if ctx == nil || h == nil || config.MinInterval <= 0 || config.MaxInterval < config.MinInterval {
		return Stats{}, fmt.Errorf("hospital, contexto e intervalos válidos obligatorios")
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	patients := h.Patients()
	results := make(chan result, len(patients))
	var wg sync.WaitGroup
	for i, p := range patients {
		wg.Add(1)
		// Arguments avoid capturing the loop variable on Go 1.21.
		go func(p *hospital.Patient, seed int64) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(seed))
			out := result{}
			defer func() { results <- out }()
			locations := []string{"Cafetería", "Pasillo del segundo piso", "Sala de espera", "Jardín", "Radiología"}
			for {
				delay := config.MinInterval
				if span := config.MaxInterval - config.MinInterval; span > 0 {
					delay += time.Duration(rng.Int63n(int64(span)))
				}
				timer := time.NewTimer(delay)
				select {
				case <-ctx.Done():
					timer.Stop()
					return
				case <-timer.C:
				}
				var err error
				switch p.State() {
				case hospital.Awake:
					err = p.SufferSleepAttack(locations[rng.Intn(len(locations))])
					if err == nil || errors.Is(err, hospital.ErrNoBed) {
						out.stats.Attacks++
					}
				case hospital.AsleepInHallway:
					_, err = h.AssignRoom(p)
				case hospital.AsleepInBed:
					err = p.WakeUp()
					if err == nil {
						out.stats.WakeUps++
					}
				}
				if errors.Is(err, hospital.ErrNoBed) {
					out.stats.BedShortages++
					continue
				}
				// A web action may change the state between the read and the command.
				if errors.Is(err, hospital.ErrAlreadyAwake) || errors.Is(err, hospital.ErrAlreadyAsleep) || errors.Is(err, hospital.ErrNotInHallway) {
					continue
				}
				if err != nil {
					out.err = err
					cancel()
					return
				}
			}
		}(p, config.Seed+int64(i))
	}
	wg.Wait()
	close(results)
	total := Stats{}
	var firstErr error
	for r := range results {
		total.Attacks += r.stats.Attacks
		total.BedShortages += r.stats.BedShortages
		total.WakeUps += r.stats.WakeUps
		if firstErr == nil {
			firstErr = r.err
		}
	}
	return total, firstErr
}
