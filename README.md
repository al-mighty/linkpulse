# LinkPulse

High-performance URL shortener with click analytics. Built with Go.

**Live:** [cheslav.space/linkpulse](https://cheslav.space/linkpulse/app)

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.23, chi router, pgx |
| Cache | Redis 7 (redirect caching) |
| Database | PostgreSQL 16 |
| Frontend | React 19, TypeScript, Tailwind |
| Deploy | Docker (15MB binary) |

## Architecture

```
POST /api/links → generate code → store in PG + cache in Redis
GET /{code}     → check Redis → fallback PG → 301 redirect → async track click (goroutine)
GET /api/links/{code}/stats → aggregate clicks by day/referer/country
```

## Getting Started

```bash
cp .env.example .env
docker compose up -d
```

Open [http://localhost:8080/app](http://localhost:8080/app)

## Author

Vyacheslav Kovalev — [cheslav.space](https://cheslav.space) · [GitHub](https://github.com/al-mighty)