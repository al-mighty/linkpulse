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
	h := handler.New(svc)

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

	r.Get("/api/health", h.Health)
	r.Post("/api/links", h.CreateLink)
	r.Get("/api/links", h.ListLinks)
	r.Get("/api/links/{code}/stats", h.GetStats)
	// Serve frontend static files
	staticDir := "/static"
	if _, err := os.Stat("frontend/dist"); err == nil {
		staticDir = "frontend/dist"
	}
	fileServer := http.FileServer(http.Dir(staticDir))
	r.Get("/app/*", http.StripPrefix("/app", fileServer).ServeHTTP)
	r.Get("/app", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, staticDir+"/index.html")
	})

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
	`)
	return err
}