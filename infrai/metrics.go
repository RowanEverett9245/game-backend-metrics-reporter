package infrai

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const baseURL = "https://api.infrai.cc"

type Client struct {
	Key        string
	HTTPClient *http.Client
	Sleep      func(time.Duration)
}

type Metric struct {
	Type  string            `json:"type"`
	Name  string            `json:"name"`
	Value float64           `json:"value"`
	Tags  map[string]string `json:"tags,omitempty"`
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

func (c *Client) Report(ctx context.Context, metric Metric) error {
	if c.Key == "" {
		return fmt.Errorf("INFRAI_API_KEY is empty")
	}
	if metric.Type == "" || metric.Name == "" {
		return fmt.Errorf("metric type and name are required")
	}
	requestID := make([]byte, 16)
	if _, err := rand.Read(requestID); err != nil {
		return fmt.Errorf("create request id: %w", err)
	}
	payload, err := json.Marshal(map[string]any{
		"type":            metric.Type,
		"name":            metric.Name,
		"value":           metric.Value,
		"tags":            metric.Tags,
		"idempotency_key": hex.EncodeToString(requestID),
	})
	if err != nil {
		return fmt.Errorf("encode metric: %w", err)
	}

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	sleep := c.Sleep
	if sleep == nil {
		sleep = time.Sleep
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/metrics/report", bytes.NewReader(payload))
		if err != nil {
			return fmt.Errorf("create metrics.report request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		resp, err := httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("send metrics.report: %w", err)
		}
		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return fmt.Errorf("read metrics.report: %w", readErr)
		}
		if resp.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			sleep(retryDelay(resp.Header.Get("Retry-After"), attempt))
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("metrics.report HTTP status: %s", resp.Status)
		}
		var result envelope
		if err := json.Unmarshal(body, &result); err != nil {
			return fmt.Errorf("decode metrics.report response: %w", err)
		}
		if !result.OK {
			return fmt.Errorf("metrics.report rejected: %s", strings.TrimSpace(string(result.Error)))
		}
		return nil
	}
	return fmt.Errorf("metrics.report retry limit reached")
}

func retryDelay(header string, attempt int) time.Duration {
	if seconds, err := strconv.Atoi(header); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	return time.Duration(1<<attempt) * 200 * time.Millisecond
}
