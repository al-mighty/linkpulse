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

type Event struct {
	ID        int64                  `json:"id"`
	Project   string                 `json:"project"`
	Name      string                 `json:"name"`
	Page      string                 `json:"page,omitempty"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	IP        string                 `json:"ip,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	Referer   string                 `json:"referer,omitempty"`
	Country   string                 `json:"country,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

type TrackEventReq struct {
	Project string                 `json:"project"`
	Name    string                 `json:"name"`
	Page    string                 `json:"page,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

type EventStats struct {
	Total        int64     `json:"total"`
	EventsByDay  []DayStat `json:"events_by_day"`
	TopEvents    []KVStat  `json:"top_events"`
	TopProjects  []KVStat  `json:"top_projects"`
	TopPages     []KVStat  `json:"top_pages"`
	TopCountries []KVStat  `json:"top_countries"`
}