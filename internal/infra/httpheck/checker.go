package httpcheck

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/kias-hack/web-watcher/internal/domain"
)

func NewChecker(client *http.Client) domain.ServiceChecker {
	return &HTTPServiceChecker{
		httpClient: client,
	}
}

type HTTPServiceChecker struct {
	httpClient *http.Client
}

func (c *HTTPServiceChecker) ServiceCheck(ctx context.Context, service *domain.Service) ([]domain.CheckResult, error) {
	logger := slog.With("component", "httpservicechecker", "service_name", service.Name, "url", service.URL)

	logger.Debug("starts service check")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, service.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed create request: %w", err)
	}

	start := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed request: %w", err)
	}
	latency := time.Since(start)
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read response body: %w", err)
	}

	logger.Debug("got service response", "status_code", resp.StatusCode, "latency", latency)

	checkInput := &domain.CheckInput{
		Response: resp,
		Latency:  latency,
		Body:     bodyBytes,
	}

	logger.Debug("runs checks")

	var result []domain.CheckResult
	for _, rule := range service.Rules {
		result = append(result, rule.Check(ctx, checkInput))
	}

	return result, nil
}
