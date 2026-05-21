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

## Event tracking

Beyond URL shortening, LinkPulse now ingests arbitrary product events
from sibling projects.

```
POST /api/events            → enqueue {project, name, page?, payload?}
GET  /api/events/stats      → aggregate by day, top events, top projects,
                              top pages, top countries
```

Embed the shared snippet to start sending pageviews and custom events:

```html
<script async src="https://cheslav.space/linkpulse/track.js"
        data-project="my-app"></script>
```

```js
cheslav.track('demo_open', { demo: 'pharma-rag' });
cheslav.track('contact_click', { source: 'main_cta' });
```

Inserts are buffered through a Go channel and processed asynchronously, so
the frontend never waits on the database. The dashboard at
[/linkpulse/app](https://cheslav.space/linkpulse/app) → Events tab plots
events per day and surfaces the top breakdowns.

## Getting Started

```bash
cp .env.example .env
docker compose up -d
```

Open [http://localhost:8080/app](http://localhost:8080/app)

## Author

Vyacheslav Kovalev — [cheslav.space](https://cheslav.space) · [GitHub](https://github.com/al-mighty)