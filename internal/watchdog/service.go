package watchdog

import (
	"net/http"
	"time"

	"github.com/kias-hack/web-watcher/internal/config"
)

type Service struct {
	url    string
	method string

	ScrapeConfig struct {
		CheckSSL             bool
		CheckSSLDate         bool
		SSLNotifyPeriod      int
		ExpectedStatusCode   int
		ExpectedBodyContains []string
	}

	config config.Service
}

type ServiceReport struct {
	Status int
	Ping   time.Duration
	Err    error
}

func checkRedirect(followRedirects bool, maxRedirects int) func(*http.Request, []*http.Request) error {
	if followRedirects && maxRedirects < 1 {
		panic("maxRedirects must be greater than or equal to 1")
	}

	return func(req *http.Request, via []*http.Request) error {
		if !followRedirects {
			return http.ErrUseLastResponse
		}

		if len(via) >= maxRedirects {
			return http.ErrUseLastResponse
		}

		return nil
	}
}

func NewService(cfg config.Service) (*Service, error) {
	return &Service{
		url: cfg.URL,
	}, nil
}

// func (s *Service) Run() *ServiceReport {
// req, err := http.NewRequest(http.MethodGet, s.url, nil)
// if err != nil {
// 	return &ServiceReport{
// 		Err: err,
// 	}
// }

// start := time.Now()

// // TODO проверить на ключение строки в ответе

// ping := time.Since(start)

// return &ServiceReport{
// 	// Status: resp.StatusCode,
// 	Ping: ping,
// }
// }
