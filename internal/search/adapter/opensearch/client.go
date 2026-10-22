package opensearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/your-org/go-minimall/internal/search/domain"
)

// Client talks to a single-node OpenSearch 2.x over HTTP.
type Client struct {
	BaseURL string
	Alias   string
	HTTP    *http.Client
}

func New(baseURL, alias string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	if alias == "" {
		alias = domain.Alias
	}
	return &Client{
		BaseURL: baseURL,
		Alias:   alias,
		HTTP:    &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) Ping(ctx context.Context) error {
	resp, err := c.do(ctx, http.MethodGet, "/", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("opensearch ping: %s", resp.Status)
	}
	return nil
}

func (c *Client) Ensure(ctx context.Context) error {
	physical := domain.PhysicalIndex
	resp, err := c.do(ctx, http.MethodHead, "/"+physical, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		body := map[string]any{
			"settings": map[string]any{
				"number_of_shards":   1,
				"number_of_replicas": 0,
			},
			"aliases": map[string]any{
				c.Alias: map[string]any{},
			},
			"mappings": map[string]any{
				"properties": map[string]any{
					"product_id":     map[string]any{"type": "keyword"},
					"category_id":    map[string]any{"type": "keyword"},
					"title":          textWithRaw(),
					"subtitle":       textWithRaw(),
					"status":         map[string]any{"type": "integer"},
					"min_price_cent": map[string]any{"type": "long"},
					"updated_at":     map[string]any{"type": "date"},
				},
			},
		}
		resp, err = c.do(ctx, http.MethodPut, "/"+physical, body)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			b, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("create index %s: %s %s", physical, resp.Status, bytes.TrimSpace(b))
		}
		return nil
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("head index %s: %s", physical, resp.Status)
	}
	aliasBody := map[string]any{
		"actions": []any{
			map[string]any{"add": map[string]any{"index": physical, "alias": c.Alias}},
		},
	}
	resp, err = c.do(ctx, http.MethodPost, "/_aliases", aliasBody)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		msg := string(b)
		if strings.Contains(msg, "already exists") {
			return nil
		}
		return fmt.Errorf("alias %s: %s %s", c.Alias, resp.Status, bytes.TrimSpace(b))
	}
	return nil
}

func textWithRaw() map[string]any {
	return map[string]any{
		"type": "text",
		"fields": map[string]any{
			"raw": map[string]any{"type": "keyword"},
		},
	}
}

func (c *Client) Upsert(ctx context.Context, doc domain.Document) error {
	body := map[string]any{
		"product_id":     doc.ProductID,
		"category_id":    doc.CategoryID,
		"title":          doc.Title,
		"subtitle":       doc.Subtitle,
		"status":         doc.Status,
		"min_price_cent": doc.MinPriceCent,
		"updated_at":     time.Now().UTC().Format(time.RFC3339Nano),
	}
	path := "/" + url.PathEscape(c.Alias) + "/_doc/" + url.PathEscape(doc.ProductID) + "?refresh=wait_for"
	resp, err := c.do(ctx, http.MethodPut, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upsert %s: %s %s", doc.ProductID, resp.Status, bytes.TrimSpace(b))
	}
	return nil
}

func (c *Client) Search(ctx context.Context, q domain.Query) ([]domain.Document, error) {
	filter := []any{
		map[string]any{"term": map[string]any{"status": 1}},
	}
	if q.CategoryID != "" {
		filter = append(filter, map[string]any{"term": map[string]any{"category_id": q.CategoryID}})
	}
	if q.PriceMin != nil || q.PriceMax != nil {
		rng := map[string]any{}
		if q.PriceMin != nil {
			rng["gte"] = *q.PriceMin
		}
		if q.PriceMax != nil {
			rng["lte"] = *q.PriceMax
		}
		filter = append(filter, map[string]any{"range": map[string]any{"min_price_cent": rng}})
	}
	boolQ := map[string]any{"filter": filter}
	if q.Text != "" {
		wild := "*" + escapeWildcard(q.Text) + "*"
		boolQ["should"] = []any{
			map[string]any{"wildcard": map[string]any{"title.raw": map[string]any{"value": wild, "case_insensitive": true}}},
			map[string]any{"wildcard": map[string]any{"subtitle.raw": map[string]any{"value": wild, "case_insensitive": true}}},
		}
		boolQ["minimum_should_match"] = 1
	}
	var sort []any
	switch q.Sort {
	case "", "new":
		sort = []any{
			map[string]any{"updated_at": map[string]string{"order": "desc"}},
			map[string]any{"product_id": map[string]string{"order": "asc"}},
		}
	case "price_asc":
		sort = []any{
			map[string]any{"min_price_cent": map[string]string{"order": "asc"}},
			map[string]any{"product_id": map[string]string{"order": "asc"}},
		}
	case "price_desc":
		sort = []any{
			map[string]any{"min_price_cent": map[string]string{"order": "desc"}},
			map[string]any{"product_id": map[string]string{"order": "asc"}},
		}
	default:
		return nil, domain.ErrBadQuery
	}
	body := map[string]any{
		"query": map[string]any{"bool": boolQ},
		"sort":  sort,
		"from":  q.Offset,
		"size":  q.Limit,
	}
	resp, err := c.do(ctx, http.MethodPost, "/"+url.PathEscape(c.Alias)+"/_search", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("search: %s %s", resp.Status, bytes.TrimSpace(raw))
	}
	var parsed struct {
		Hits struct {
			Hits []struct {
				Source domain.Document `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, err
	}
	out := make([]domain.Document, 0, len(parsed.Hits.Hits))
	for _, hit := range parsed.Hits.Hits {
		out = append(out, hit.Source)
	}
	return out, nil
}

func escapeWildcard(s string) string {
	return strings.NewReplacer(`\`, `\\`, `*`, `\*`, `?`, `\?`).Replace(s)
}

func (c *Client) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, r)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.HTTP.Do(req)
}
