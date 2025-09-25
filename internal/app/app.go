package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	bshttp "broadcast_server/internal/adapters/http"
	"broadcast_server/internal/domain"
	"broadcast_server/internal/events"
)

type Config struct {
	Addr string `json:"addr,omitempty"`
}

func Run(cfg Config) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	hub := domain.NewHub()
	go hub.Run(ctx)

	evBus := events.NewSubject[events.Event](128, 0)
	defer evBus.Close()

	sub := evBus.Subscribe()
	go func() {
		for ev := range sub.C {
			switch ev.Type {
			case events.EventJoined:
				log.Printf("event=joined sender=%s", ev.Sender)
			case events.EventLeft:
				log.Printf("event=left sender=%s", ev.Sender)
			case events.EventBroadcast:
				log.Printf("event=broadcast sender=%s size=%d", ev.Sender, len(ev.Payload))
			default:
				log.Printf("event=%s sender=%s size=%d", ev.Type, ev.Sender, len(ev.Payload))
			}
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		bshttp.ServeWS(ctx, hub, evBus, w, r)
	})

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("http shutdown error: %v", err)
		}
	}()

	log.Printf("listening on %s", cfg.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
