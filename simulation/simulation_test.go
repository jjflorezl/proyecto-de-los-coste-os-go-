package simulation

import (
	"context"
	"costenos-narcolepsia/internal/scenario"
	"testing"
	"time"
)

func TestRunConcurrentSimulation(t *testing.T) {
	h, err := scenario.New()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()
	stats, err := Run(ctx, h, Config{MinInterval: time.Millisecond, MaxInterval: 3 * time.Millisecond, Seed: 9})
	if err != nil {
		t.Fatal(err)
	}
	if stats.Attacks == 0 || stats.WakeUps == 0 || len(h.History()) != stats.Attacks {
		t.Fatalf("estadísticas: %+v", stats)
	}
	seen := map[string]bool{}
	for _, room := range h.Snapshot().Rooms {
		if len(room.OccupantIDs) > room.Capacity {
			t.Fatal("sobreocupación")
		}
		for _, id := range room.OccupantIDs {
			if seen[id] {
				t.Fatal("doble ocupación")
			}
			seen[id] = true
		}
	}
}

func TestInvalidConfiguration(t *testing.T) {
	h, err := scenario.New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Run(context.Background(), h, Config{}); err == nil {
		t.Fatal("intervalos inválidos aceptados")
	}
}
