package service

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/al-mighty/linkpulse/internal/models"
	"github.com/al-mighty/linkpulse/internal/repo"
)

const serviceTimeout = 5 * time.Second

type EventService struct {
	db      *repo.Postgres
	eventCh chan *models.Event
}

func NewEventService(db *repo.Postgres) *EventService {
	s := &EventService{
		db:      db,
		eventCh: make(chan *models.Event, 1000),
	}
	go s.processEvents()
	return s
}

// Track validates and enqueues an event for async insert. Returns fast so the
// browser can keep moving. The buffered channel absorbs short bursts; if it
// overflows we drop the event and log (we'd rather lose telemetry than block
// the request path).
func (s *EventService) Track(ev *models.Event) error {
	if ev.Project == "" || ev.Name == "" {
		return errors.New("project and name are required")
	}
	if len(ev.Project) > 64 {
		ev.Project = ev.Project[:64]
	}
	if len(ev.Name) > 128 {
		ev.Name = ev.Name[:128]
	}
	if len(ev.Page) > 1024 {
		ev.Page = ev.Page[:1024]
	}
	ev.Project = strings.TrimSpace(ev.Project)
	ev.Name = strings.TrimSpace(ev.Name)

	select {
	case s.eventCh <- ev:
		return nil
	default:
		log.Printf("event: channel full, dropping event %s/%s", ev.Project, ev.Name)
		return errors.New("event queue full")
	}
}

func (s *EventService) processEvents() {
	for ev := range s.eventCh {
		ctx, cancel := context.WithTimeout(context.Background(), serviceTimeout)
		if err := s.db.InsertEvent(ctx, ev); err != nil {
			log.Printf("event: insert failed: %v", err)
		}
		cancel()
	}
}

func (s *EventService) Stats(ctx context.Context, project string, sinceDays int) (*models.EventStats, error) {
	return s.db.EventStats(ctx, project, sinceDays)
}
