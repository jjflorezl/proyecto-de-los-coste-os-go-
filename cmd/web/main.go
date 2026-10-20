package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"costenos-narcolepsia/internal/scenario"
	"costenos-narcolepsia/web"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "dirección local de la interfaz")
	flag.Parse()
	if err := run(*addr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(addr string) error {
	h, err := scenario.New()
	if err != nil {
		return err
	}
	app, err := web.New(h)
	if err != nil {
		return err
	}
	defer app.Close()
	server := &http.Server{Addr: addr, Handler: app.Handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	result := make(chan error, 1)
	go func() { result <- server.ListenAndServe() }()
	fmt.Printf("Interfaz del hospital: http://%s\nPresiona Ctrl+C para cerrar.\n", addr)
	select {
	case err := <-result:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return server.Shutdown(shutdown)
	}
}
