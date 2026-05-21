package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/al-mighty/linkpulse/internal/handler"
	"github.com/al-mighty/linkpulse/internal/repo"
	"github.com/al-mighty/linkpulse/internal/service"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {
	godotenv.Load()

	ctx := context.Background()

	// Postgres
	pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("postgres:", err)
	}
	defer pool.Close()

	// Run migrations
	if err := migrate(ctx, pool); err != nil {
		log.Fatal("migrate:", err)
	}

	// Redis
	opt, err := redis.ParseURL(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Fatal("redis url:", err)
	}
	rdb := redis.NewClient(opt)
	defer rdb.Close()

	// Services
	pg := repo.NewPostgres(pool)
	cache := repo.NewRedis(rdb)
	svc := service.NewLinkService(pg, cache, os.Getenv("BASE_URL"))
	events := service.NewEventService(pg)
	h := handler.New(svc, events)

	// Router
	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type"},
	}))

	// Static files dir
	staticDir := "/static"
	if _, err := os.Stat("frontend/dist"); err == nil {
		staticDir = "frontend/dist"
	}

	r.Get("/api/health", h.Health)
	r.Post("/api/links", h.CreateLink)
	r.Get("/api/links", h.ListLinks)
	r.Get("/api/links/{code}/stats", h.GetStats)
	r.Post("/api/events", h.TrackEvent)
	r.Get("/api/events/stats", h.EventStats)

	// Frontend: serve index.html for /app, static assets for /app/*
	r.Get("/app", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, staticDir+"/index.html")
	})
	r.Get("/app/*", func(w http.ResponseWriter, r *http.Request) {
		// Try static file first, fallback to index.html (SPA)
		path := staticDir + r.URL.Path[4:] // strip "/app"
		if _, err := os.Stat(path); err == nil {
			http.StripPrefix("/app", http.FileServer(http.Dir(staticDir))).ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, staticDir+"/index.html")
	})

	// Also serve assets from root /assets/ (when behind reverse proxy stripping prefix)
	r.Get("/assets/*", http.FileServer(http.Dir(staticDir)).ServeHTTP)

	// Root serves frontend
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, staticDir+"/index.html")
	})

	// Short link redirect — must be last
	r.Get("/{code}", h.Redirect)

	// Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("LinkPulse listening on :%s", port)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	shutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	srv.Shutdown(shutCtx)
}

func migrate(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS links (
			id BIGSERIAL PRIMARY KEY,
			code VARCHAR(10) UNIQUE NOT NULL,
			url TEXT NOT NULL,
			clicks BIGINT DEFAULT 0,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_links_code ON links(code);

		CREATE TABLE IF NOT EXISTS clicks (
			id BIGSERIAL PRIMARY KEY,
			link_id BIGINT REFERENCES links(id) ON DELETE CASCADE,
			ip VARCHAR(45),
			user_agent TEXT,
			referer TEXT,
			country VARCHAR(2),
			created_at TIMESTAMPTZ DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_clicks_link_id ON clicks(link_id);
		CREATE INDEX IF NOT EXISTS idx_clicks_created ON clicks(created_at);

		CREATE TABLE IF NOT EXISTS events (
			id BIGSERIAL PRIMARY KEY,
			project VARCHAR(64) NOT NULL,
			name VARCHAR(128) NOT NULL,
			page TEXT,
			payload JSONB,
			ip VARCHAR(45),
			user_agent TEXT,
			referer TEXT,
			country VARCHAR(2),
			created_at TIMESTAMPTZ DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_events_project ON events(project);
		CREATE INDEX IF NOT EXISTS idx_events_created ON events(created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_events_proj_name ON events(project, name);
	`)
	return err
}