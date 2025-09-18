package thecatapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type HTTPClient struct {
	baseURL string
	apiKey  string
	http    *http.Client

	mu        sync.RWMutex
	cacheTill time.Time
	cache     []Breed
	ttl       time.Duration
}

func NewHTTP(ttl time.Duration) *HTTPClient {
	return &HTTPClient{
		baseURL: "https://api.thecatapi.com/v1",
		apiKey:  os.Getenv("THECATAPI_KEY"),
		http:    &http.Client{Timeout: 10 * time.Second},
		ttl:     ttl,
	}
}

func (c *HTTPClient) ListBreeds() ([]Breed, error) {
	c.mu.RLock()
	if time.Now().Before(c.cacheTill) && c.cache != nil {
		b := c.cache
		c.mu.RUnlock()
		return b, nil
	}
	c.mu.RUnlock()

	req, _ := http.NewRequest(http.MethodGet, c.baseURL+"/breeds", nil)
	if c.apiKey != "" {
		req.Header.Set("x-api-key", c.apiKey)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return nil, errors.New("catapi: non-200")
	}

	var list []Breed
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.cache = list
	c.cacheTill = time.Now().Add(c.ttl)
	c.mu.Unlock()
	return list, nil
}

func (c *HTTPClient) ValidateBreed(nameOrID string) (bool, error) {
	list, err := c.ListBreeds()
	if err != nil {
		return false, err
	}
	in := strings.ToLower(strings.TrimSpace(nameOrID))
	for _, b := range list {
		if strings.ToLower(b.Name) == in || strings.ToLower(b.ID) == in {
			return true, nil
		}
	}
	return false, nil
}
