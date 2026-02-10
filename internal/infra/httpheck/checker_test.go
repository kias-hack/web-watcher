package httpcheck

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kias-hack/web-watcher/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestHTTPServiceChecker(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><head><title>Test</title></head><body>Test</body></html>"))
	}))
	defer server.Close()

	checker := HTTPServiceChecker{
		httpClient: http.DefaultClient,
	}

	result, err := checker.ServiceCheck(t.Context(), &domain.Service{
		URL: server.URL,
		Rules: []domain.CheckRule{
			domain.NewStatusCodeRule(http.StatusOK),
			domain.NewLatencyRule(150),
			domain.NewBodyMatchRule("<title>Test</title>"),
		},
	})

	assert.NoError(t, err)
	for _, res := range result {
		t.Log(res.RuleType, res.Message)
		assert.Equal(t, domain.OK, res.OK)
	}
}
