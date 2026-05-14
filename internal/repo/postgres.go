package repo

import (
	"context"
	"fmt"

	"github.com/al-mighty/linkpulse/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

func (p *Postgres) CreateLink(ctx context.Context, code, url string) (*models.Link, error) {
	link := &models.Link{}
	err := p.pool.QueryRow(ctx,
		`INSERT INTO links (code, url) VALUES ($1, $2) RETURNING id, code, url, clicks, created_at`,
		code, url,
	).Scan(&link.ID, &link.Code, &link.URL, &link.Clicks, &link.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create link: %w", err)
	}
	return link, nil
}

func (p *Postgres) GetLinkByCode(ctx context.Context, code string) (*models.Link, error) {
	link := &models.Link{}
	err := p.pool.QueryRow(ctx,
		`SELECT id, code, url, clicks, created_at FROM links WHERE code = $1`,
		code,
	).Scan(&link.ID, &link.Code, &link.URL, &link.Clicks, &link.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get link: %w", err)
	}
	return link, nil
}

func (p *Postgres) IncrClicks(ctx context.Context, id int64) error {
	_, err := p.pool.Exec(ctx, `UPDATE links SET clicks = clicks + 1 WHERE id = $1`, id)
	return err
}

func (p *Postgres) SaveClick(ctx context.Context, click *models.Click) error {
	_, err := p.pool.Exec(ctx,
		`INSERT INTO clicks (link_id, ip, user_agent, referer, country) VALUES ($1, $2, $3, $4, $5)`,
		click.LinkID, click.IP, click.UserAgent, click.Referer, click.Country,
	)
	return err
}

func (p *Postgres) ListLinks(ctx context.Context) ([]models.Link, error) {
	rows, err := p.pool.Query(ctx,
		`SELECT id, code, url, clicks, created_at FROM links ORDER BY created_at DESC LIMIT 50`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []models.Link
	for rows.Next() {
		var l models.Link
		if err := rows.Scan(&l.ID, &l.Code, &l.URL, &l.Clicks, &l.CreatedAt); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	return links, nil
}

func (p *Postgres) GetLinkStats(ctx context.Context, code string) (*models.LinkStats, error) {
	link, err := p.GetLinkByCode(ctx, code)
	if err != nil {
		return nil, err
	}

	stats := &models.LinkStats{Link: *link}

	// Clicks by day (last 30 days)
	rows, err := p.pool.Query(ctx, `
		SELECT DATE(created_at) as day, COUNT(*)
		FROM clicks WHERE link_id = $1 AND created_at > NOW() - INTERVAL '30 days'
		GROUP BY day ORDER BY day`, link.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var d models.DayStat
		if err := rows.Scan(&d.Date, &d.Count); err != nil {
			return nil, err
		}
		stats.ClicksByDay = append(stats.ClicksByDay, d)
	}

	// Top referers
	rows2, err := p.pool.Query(ctx, `
		SELECT COALESCE(NULLIF(referer,''), 'direct') as ref, COUNT(*) as cnt
		FROM clicks WHERE link_id = $1 GROUP BY ref ORDER BY cnt DESC LIMIT 10`, link.ID)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var s models.KVStat
		if err := rows2.Scan(&s.Key, &s.Count); err != nil {
			return nil, err
		}
		stats.TopReferers = append(stats.TopReferers, s)
	}

	// Top countries
	rows3, err := p.pool.Query(ctx, `
		SELECT COALESCE(NULLIF(country,''), 'unknown') as c, COUNT(*) as cnt
		FROM clicks WHERE link_id = $1 GROUP BY c ORDER BY cnt DESC LIMIT 10`, link.ID)
	if err != nil {
		return nil, err
	}
	defer rows3.Close()
	for rows3.Next() {
		var s models.KVStat
		if err := rows3.Scan(&s.Key, &s.Count); err != nil {
			return nil, err
		}
		stats.TopCountries = append(stats.TopCountries, s)
	}

	return stats, nil
}