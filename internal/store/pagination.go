package store

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PaginatedFeedQuery struct {
	Limit  int        `json:"limit" validate:"gte=0,lte=100"`
	Offset int        `json:"offset" validate:"gte=0,lte=100"`
	Sort   string     `json:"sort" validate:"oneof=asc desc"`
	Tags   []string   `json:"tags" validate:"max=8"`
	Search string     `json:"search" validate:"max=100"`
	Since  *time.Time `json:"since"`
	Until  *time.Time `json:"until"`
}

func (fq PaginatedFeedQuery) Parse(r *http.Request) (PaginatedFeedQuery, error) {
	qs := r.URL.Query()

	fq.Limit = 20
	fq.Offset = 0
	fq.Sort = "desc"

	if limit := qs.Get("limit"); limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil {
			return PaginatedFeedQuery{}, fmt.Errorf("invalid limit: %w", err)
		}
		fq.Limit = l
	}

	if offset := qs.Get("offset"); offset != "" {
		o, err := strconv.Atoi(offset)
		if err != nil {
			return PaginatedFeedQuery{}, fmt.Errorf("invalid offset: %w", err)
		}
		fq.Offset = o
	}

	if sort := qs.Get("sort"); sort != "" {
		fq.Sort = sort
	}

	if tags := qs.Get("tags"); tags != "" {
		fq.Tags = strings.Split(tags, ",")
	}

	if search := qs.Get("search"); search != "" {
		fq.Search = search
	}

	if since := qs.Get("since"); since != "" {
		t, err := parseTime(since)
		if err != nil {
			return PaginatedFeedQuery{}, fmt.Errorf("invalid 'since' timestamp (use RFC3339): %w", err)
		}
		fq.Since = t
	}

	if until := qs.Get("until"); until != "" {
		t, err := parseTime(until)
		if err != nil {
			return PaginatedFeedQuery{}, fmt.Errorf("invalid 'until' timestamp (use RFC3339): %w", err)
		}
		fq.Until = t
	}

	return fq, nil
}

func parseTime(s string) (*time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return &t, nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		t = t.UTC()
		return &t, nil
	}
	return nil, fmt.Errorf("unsupported time format: expected RFC3339 or YYYY-MM-DD")
}
