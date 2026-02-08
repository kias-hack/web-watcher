package watchdog

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScraper_ScrapeCases(t *testing.T) {
	testCases := []struct {
		name                string
		statusCode          int
		expectedCode        int
		responseBody        string
		expectedBodyStrings []string
		hasError            bool
	}{
		{
			name:                "некорректный статус ответа сервера",
			statusCode:          500,
			expectedCode:        200,
			responseBody:        "{\"status\": \"Internal server error\"}",
			expectedBodyStrings: nil,
			hasError:            true,
		},
		{
			name:                "отсутствуют строки в теле",
			statusCode:          200,
			expectedCode:        200,
			responseBody:        "<html><head>Test</head><body>Content body</body></html>",
			expectedBodyStrings: []string{"<head>Not test</head>"},
			hasError:            true,
		},
		{
			name:                "присутствуют строки в ответе сервера",
			statusCode:          200,
			expectedCode:        200,
			responseBody:        "<html><head>Test</head><body>Content body</body></html>",
			expectedBodyStrings: []string{"<head>Test</head>", "Content body"},
			hasError:            false,
		},
		{
			name:                "корректный ответ сервера",
			statusCode:          200,
			expectedCode:        200,
			responseBody:        "<html><head>Test</head><body>Content body</body></html>",
			expectedBodyStrings: nil,
			hasError:            false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			client := &http.Client{
				CheckRedirect: checkRedirect(false, 1),
				Transport:     http.DefaultTransport,
				Timeout:       1 * time.Second,
			}

			scraper := Scraper{
				client: client,
			}

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond)
				w.WriteHeader(testCase.statusCode)
				w.Write([]byte(testCase.responseBody))
			}))
			defer server.Close()

			result := scraper.Scrape(t.Context(), &Service{
				url:    server.URL,
				method: http.MethodGet,
				ScrapeConfig: struct {
					CheckSSL             bool
					CheckSSLDate         bool
					SSLNotifyPeriod      int
					ExpectedStatusCode   int
					ExpectedBodyContains []string
				}{
					CheckSSL:             false,
					CheckSSLDate:         false,
					SSLNotifyPeriod:      0,
					ExpectedStatusCode:   testCase.expectedCode,
					ExpectedBodyContains: testCase.expectedBodyStrings,
				},
			})

			assert.Equal(t, testCase.statusCode, result.StatusCode)
			if testCase.hasError {
				assert.Error(t, result.Err)
			} else {
				assert.NoError(t, result.Err)
			}
		})
	}
}

func TestScraper_ScrapeTimeout(t *testing.T) {
	client := &http.Client{
		CheckRedirect: checkRedirect(false, 1),
		Transport:     http.DefaultTransport,
		Timeout:       100 * time.Millisecond,
	}

	scraper := Scraper{
		client: client,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	result := scraper.Scrape(t.Context(), &Service{
		url:    server.URL,
		method: http.MethodGet,
		ScrapeConfig: struct {
			CheckSSL             bool
			CheckSSLDate         bool
			SSLNotifyPeriod      int
			ExpectedStatusCode   int
			ExpectedBodyContains []string
		}{
			CheckSSL:             false,
			CheckSSLDate:         false,
			SSLNotifyPeriod:      0,
			ExpectedStatusCode:   200,
			ExpectedBodyContains: nil,
		},
	})

	require.Error(t, result.Err)
	assert.True(t, errors.Is(result.Err, context.DeadlineExceeded),
		"ожидалась ошибка таймаута (DeadlineExceeded), получено: %v", result.Err)
}

func TestScraper_ScrapeTLSInvalidDNS(t *testing.T) {
	server := createTlsServer([]string{"example.com"}, time.Now(), time.Now().Add(+time.Minute))
	defer server.Close()

	serverURL, _ := url.Parse(server.URL)

	client := &http.Client{
		CheckRedirect: checkRedirect(false, 1),
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return (&net.Dialer{
					Timeout: 500 * time.Millisecond,
				}).DialContext(ctx, network, net.JoinHostPort(serverURL.Hostname(), serverURL.Port()))
			},
		},
		Timeout: 100 * time.Millisecond,
	}

	scraper := Scraper{
		client: client,
	}

	result := scraper.Scrape(t.Context(), &Service{
		url:    "https://server.com/",
		method: http.MethodGet,
		ScrapeConfig: struct {
			CheckSSL             bool
			CheckSSLDate         bool
			SSLNotifyPeriod      int
			ExpectedStatusCode   int
			ExpectedBodyContains []string
		}{
			CheckSSL:             true,
			CheckSSLDate:         true,
			SSLNotifyPeriod:      1,
			ExpectedStatusCode:   200,
			ExpectedBodyContains: nil,
		},
	})

	assert.ErrorContains(t, result.Err, "x509: certificate is valid for")
}

func TestScraper_ScrapeTLSValidDNS(t *testing.T) {
	server := createTlsServer([]string{"example.com"}, time.Now(), time.Now().Add(+time.Minute))
	defer server.Close()

	serverURL, _ := url.Parse(server.URL)

	client := &http.Client{
		CheckRedirect: checkRedirect(false, 1),
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return (&net.Dialer{
					Timeout: 500 * time.Millisecond,
				}).DialContext(ctx, network, net.JoinHostPort(serverURL.Hostname(), serverURL.Port()))
			},
		},
		Timeout: 100 * time.Millisecond,
	}

	scraper := Scraper{
		client: client,
	}

	result := scraper.Scrape(t.Context(), &Service{
		url:    "https://example.com/",
		method: http.MethodGet,
		ScrapeConfig: struct {
			CheckSSL             bool
			CheckSSLDate         bool
			SSLNotifyPeriod      int
			ExpectedStatusCode   int
			ExpectedBodyContains []string
		}{
			CheckSSL:             true,
			CheckSSLDate:         true,
			SSLNotifyPeriod:      1,
			ExpectedStatusCode:   200,
			ExpectedBodyContains: nil,
		},
	})

	assert.NoError(t, result.Err)
}

func TestScraper_ScrapeTLSCertOverdueComing(t *testing.T) {
	server := createTlsServer([]string{"example.com"}, time.Now().Add(-time.Hour), time.Now().Add(24*4*time.Hour))
	defer server.Close()

	serverURL, _ := url.Parse(server.URL)

	client := &http.Client{
		CheckRedirect: checkRedirect(false, 1),
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return (&net.Dialer{
					Timeout: 500 * time.Millisecond,
				}).DialContext(ctx, network, net.JoinHostPort(serverURL.Hostname(), serverURL.Port()))
			},
		},
		Timeout: 100 * time.Millisecond,
	}

	scraper := Scraper{
		client: client,
	}

	result := scraper.Scrape(t.Context(), &Service{
		url:    "https://example.com/",
		method: http.MethodGet,
		ScrapeConfig: struct {
			CheckSSL             bool
			CheckSSLDate         bool
			SSLNotifyPeriod      int
			ExpectedStatusCode   int
			ExpectedBodyContains []string
		}{
			CheckSSL:             true,
			CheckSSLDate:         true,
			SSLNotifyPeriod:      3,
			ExpectedStatusCode:   200,
			ExpectedBodyContains: nil,
		},
	})

	assert.True(t, result.NeedSSLNotify)
}

func TestScraper_ScrapeTLSCertOverdue(t *testing.T) {
	server := createTlsServer([]string{"example.com"}, time.Now().Add(-time.Hour), time.Now().Add(-time.Minute))
	defer server.Close()

	serverURL, _ := url.Parse(server.URL)

	client := &http.Client{
		CheckRedirect: checkRedirect(false, 1),
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return (&net.Dialer{
					Timeout: 500 * time.Millisecond,
				}).DialContext(ctx, network, net.JoinHostPort(serverURL.Hostname(), serverURL.Port()))
			},
		},
		Timeout: 100 * time.Millisecond,
	}

	scraper := Scraper{
		client: client,
	}

	result := scraper.Scrape(t.Context(), &Service{
		url:    "https://example.com/",
		method: http.MethodGet,
		ScrapeConfig: struct {
			CheckSSL             bool
			CheckSSLDate         bool
			SSLNotifyPeriod      int
			ExpectedStatusCode   int
			ExpectedBodyContains []string
		}{
			CheckSSL:             true,
			CheckSSLDate:         true,
			SSLNotifyPeriod:      1,
			ExpectedStatusCode:   200,
			ExpectedBodyContains: nil,
		},
	})

	assert.ErrorContains(t, result.Err, "сертификат просрочен -")
	assert.True(t, result.NeedSSLNotify)
}

func TestScraper_ScrapeTLSCertNotBeforeError(t *testing.T) {
	server := createTlsServer([]string{"example.com"}, time.Now().Add(time.Hour), time.Now().Add(24*time.Hour))
	defer server.Close()

	serverURL, _ := url.Parse(server.URL)

	client := &http.Client{
		CheckRedirect: checkRedirect(false, 1),
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return (&net.Dialer{
					Timeout: 500 * time.Millisecond,
				}).DialContext(ctx, network, net.JoinHostPort(serverURL.Hostname(), serverURL.Port()))
			},
		},
		Timeout: 100 * time.Millisecond,
	}

	scraper := Scraper{
		client: client,
	}

	result := scraper.Scrape(t.Context(), &Service{
		url:    "https://example.com/",
		method: http.MethodGet,
		ScrapeConfig: struct {
			CheckSSL             bool
			CheckSSLDate         bool
			SSLNotifyPeriod      int
			ExpectedStatusCode   int
			ExpectedBodyContains []string
		}{
			CheckSSL:             true,
			CheckSSLDate:         true,
			SSLNotifyPeriod:      1,
			ExpectedStatusCode:   200,
			ExpectedBodyContains: nil,
		},
	})

	assert.ErrorContains(t, result.Err, "ошибка начала действия сертификата -")
}

// TODO проверка сертификата, время сертификата и проверка доменного имени сертификата

func createTlsServer(dnsNames []string, notBefore time.Time, notAfter time.Time) *httptest.Server {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	certDer, key := generateCert(dnsNames, notBefore, notAfter)
	server.Config.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{certDer},
				PrivateKey:  key,
			},
		},
	}
	server.TLS = &tls.Config{
		Certificates: []tls.Certificate{
			{
				Certificate: [][]byte{certDer},
				PrivateKey:  key,
			},
		},
	}

	server.StartTLS()

	return server
}

func generateCert(dnsNames []string, notBefore time.Time, notAfter time.Time) ([]byte, crypto.PrivateKey) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		DNSNames:     dnsNames,
		KeyUsage:     x509.KeyUsageContentCommitment,
	}

	certDer, _ := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)

	return certDer, key
}
