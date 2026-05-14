package models

import "time"

type Link struct {
	ID        int64     `json:"id"`
	Code      string    `json:"code"`
	URL       string    `json:"url"`
	Clicks    int64     `json:"clicks"`
	CreatedAt time.Time `json:"created_at"`
}

type Click struct {
	ID        int64     `json:"id"`
	LinkID    int64     `json:"link_id"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Referer   string    `json:"referer"`
	Country   string    `json:"country"`
	CreatedAt time.Time `json:"created_at"`
}

type LinkStats struct {
	Link
	ClicksByDay []DayStat `json:"clicks_by_day"`
	TopReferers []KVStat  `json:"top_referers"`
	TopCountries []KVStat `json:"top_countries"`
}

type DayStat struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type KVStat struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

type CreateLinkReq struct {
	URL string `json:"url"`
}

type CreateLinkResp struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
	URL      string `json:"url"`
}