package service

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/linkcheck/pkg/types"
)

// Checker сервис для проверки доступности ссылок
type Checker struct {
	client *http.Client
}

// NewChecker создает новый сервис проверки ссылок
func NewChecker() *Checker {
	return &Checker{
		client: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// CheckLink проверяет доступность одной ссылки
func (c *Checker) CheckLink(ctx context.Context, link string) types.LinkStatus {
	normalizedURL := c.normalizeURL(link)

	req, err := http.NewRequestWithContext(ctx, "GET", normalizedURL, nil)
	if err != nil {
		return types.StatusNotAvailable
	}

	req.Header.Set("User-Agent", "LinkCheck/1.0")

	resp, err := c.client.Do(req)
	if err != nil {
		return types.StatusNotAvailable
	}
	defer resp.Body.Close()

	if resp.StatusCode < 500 {
		return types.StatusAvailable
	}

	return types.StatusNotAvailable
}

// normalizeURL добавляет протокол к ссылке, если его нет
func (c *Checker) normalizeURL(link string) string {
	link = strings.TrimSpace(link)

	if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
		return link
	}

	httpsURL := "https://" + link
	if _, err := url.Parse(httpsURL); err == nil {
		return httpsURL
	}

	return "http://" + link
}

// CheckLinks проверяет доступность нескольких ссылок
func (c *Checker) CheckLinks(ctx context.Context, links []string) map[string]types.LinkStatus {
	results := make(map[string]types.LinkStatus, len(links))

	for _, link := range links {
		status := c.CheckLink(ctx, link)
		results[link] = status
	}

	return results
}
