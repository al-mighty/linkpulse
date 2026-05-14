package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
	"strings"

	"github.com/al-mighty/linkpulse/internal/models"
	"github.com/al-mighty/linkpulse/internal/repo"
	"github.com/redis/go-redis/v9"
)

type LinkService struct {
	db      *repo.Postgres
	cache   *repo.Redis
	baseURL string
	clickCh chan *models.Click
}

func NewLinkService(db *repo.Postgres, cache *repo.Redis, baseURL string) *LinkService {
	s := &LinkService{
		db:      db,
		cache:   cache,
		baseURL: strings.TrimRight(baseURL, "/"),
		clickCh: make(chan *models.Click, 1000),
	}
	go s.processClicks()
	return s
}

func (s *LinkService) Create(ctx context.Context, url string) (*models.CreateLinkResp, error) {
	code := generateCode()
	link, err := s.db.CreateLink(ctx, code, url)
	if err != nil {
		return nil, err
	}

	_ = s.cache.CacheURL(ctx, code, url)

	return &models.CreateLinkResp{
		Code:     link.Code,
		ShortURL: s.baseURL + "/" + link.Code,
		URL:      link.URL,
	}, nil
}

func (s *LinkService) Resolve(ctx context.Context, code string) (string, int64, error) {
	// Try cache first
	if url, err := s.cache.GetCachedURL(ctx, code); err == nil {
		link, _ := s.db.GetLinkByCode(ctx, code)
		if link != nil {
			return url, link.ID, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		log.Printf("redis error: %v", err)
	}

	// Fallback to DB
	link, err := s.db.GetLinkByCode(ctx, code)
	if err != nil {
		return "", 0, err
	}

	_ = s.cache.CacheURL(ctx, code, link.URL)
	return link.URL, link.ID, nil
}

func (s *LinkService) TrackClick(linkID int64, ip, userAgent, referer string) {
	s.clickCh <- &models.Click{
		LinkID:    linkID,
		IP:        ip,
		UserAgent: userAgent,
		Referer:   referer,
	}
}

func (s *LinkService) processClicks() {
	for click := range s.clickCh {
		ctx := context.Background()
		if err := s.db.IncrClicks(ctx, click.LinkID); err != nil {
			log.Printf("incr clicks error: %v", err)
		}
		if err := s.db.SaveClick(ctx, click); err != nil {
			log.Printf("save click error: %v", err)
		}
	}
}

func (s *LinkService) List(ctx context.Context) ([]models.Link, error) {
	return s.db.ListLinks(ctx)
}

func (s *LinkService) Stats(ctx context.Context, code string) (*models.LinkStats, error) {
	return s.db.GetLinkStats(ctx, code)
}

func generateCode() string {
	b := make([]byte, 6)
	rand.Read(b)
	code := base64.RawURLEncoding.EncodeToString(b)
	if len(code) > 7 {
		code = code[:7]
	}
	return code
}