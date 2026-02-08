package watchdog

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ScrapeResult struct {
	URL           string
	StatusCode    int
	Latency       time.Duration
	Err           error
	NeedSSLNotify bool
}

type Scraper struct {
	client *http.Client
}

func (s *Scraper) Scrape(ctx context.Context, service *Service) ScrapeResult {
	result := ScrapeResult{}
	// TODO сделать запрос к по url
	// засечь время
	// получить информацию о сертификате
	// проверить вхождения в теле
	// проверить все правила

	req, err := http.NewRequestWithContext(ctx, service.method, service.url, nil)
	if err != nil {
		result.Err = err
		return result
	}

	start := time.Now()
	response, err := s.client.Do(req)
	result.Latency = time.Since(start)
	if err != nil {
		result.Err = err
		return result
	}

	var errList []error

	if service.ScrapeConfig.CheckSSL {
		if response.TLS == nil {
			errList = append(errList, fmt.Errorf("отсутствует информация о TLS соединении"))
		} else if len(response.TLS.PeerCertificates) == 0 {
			errList = append(errList, fmt.Errorf("сертификаты отсутствуют в ответе"))
		} else {
			cert := response.TLS.PeerCertificates[0]

			if err := cert.VerifyHostname(response.Request.URL.Hostname()); err != nil {
				errList = append(errList, err)
			}

			left := time.Until(cert.NotAfter)
			leftInDays := int(left.Hours() / 24)

			if time.Until(cert.NotBefore) > 0 {
				errList = append(errList, fmt.Errorf("ошибка начала действия сертификата - %s", cert.NotBefore))
			}

			if left < 0 {
				errList = append(errList, fmt.Errorf("сертификат просрочен - %s", cert.NotAfter))
			}

			if service.ScrapeConfig.CheckSSLDate && leftInDays <= service.ScrapeConfig.SSLNotifyPeriod {
				result.NeedSSLNotify = true
			}
		}
	}

	if response.StatusCode != service.ScrapeConfig.ExpectedStatusCode {
		errList = append(errList, fmt.Errorf("ошибка кода ответа сервера, ожидается %d получено %d", service.ScrapeConfig.ExpectedStatusCode, response.StatusCode))
	}
	result.StatusCode = response.StatusCode

	body, err := io.ReadAll(response.Body)
	if err != nil {
		result.Err = err
		return result
	}
	builder := strings.Builder{}
	builder.Write(body)
	bodyStr := builder.String()

	for _, expectedStr := range service.ScrapeConfig.ExpectedBodyContains {
		if !strings.Contains(bodyStr, expectedStr) {
			errList = append(errList, fmt.Errorf("Строка \"%s\" не найдена в теле ответа сервера", expectedStr))
		}
	}

	if len(errList) > 0 {
		result.Err = errors.Join(errList...)
	}

	return result
}

// func checkCertificateDomain(cert *x509.Certificate, domain string) {
// 	cert.DNSNames
// }
