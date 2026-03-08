package weather

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

type stubHTTPClient struct {
	do func(req *http.Request) (*http.Response, error)
}

func (c stubHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.do(req)
}

func TestNewServiceRequiresAPIKey(t *testing.T) {
	if _, err := NewService(nil, ""); err == nil {
		t.Fatal("NewService() error = nil, want non-nil")
	}
}

func TestNewServiceUsesDefaultHTTPClient(t *testing.T) {
	service, err := NewService(nil, "test-key")
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	if service.client == nil {
		t.Fatal("service.client = nil, want non-nil")
	}
}

func TestFormatCurrentWeather(t *testing.T) {
	var requests []*http.Request

	service, err := NewService(stubHTTPClient{
		do: func(req *http.Request) (*http.Response, error) {
			requests = append(requests, req)

			switch req.URL.Query().Get("q") {
			case "Astrakhan":
				return jsonResponse(http.StatusOK, weatherPayload("Astrakhan City", 5.5, 3.2, "Cloudy")), nil
			case "Montreal":
				return jsonResponse(http.StatusOK, weatherPayload("Montreal", -2.0, -7.1, "Clear")), nil
			case "Seattle":
				return jsonResponse(http.StatusOK, weatherPayload("Seattle", 9.3, 7.8, "Light rain")), nil
			default:
				return nil, fmt.Errorf("unexpected city %q", req.URL.Query().Get("q"))
			}
		},
	}, "test-key")
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	got, err := service.FormatCurrentWeather(context.Background())
	if err != nil {
		t.Fatalf("FormatCurrentWeather() error = %v", err)
	}

	want := strings.Join([]string{
		"Astrakhan City: 6\u00b0 (3\u00b0) / Cloudy",
		"Montreal: -2\u00b0 (-7\u00b0) / Clear",
		"Seattle: 9\u00b0 (8\u00b0) / Light rain",
	}, "\n")

	if got != want {
		t.Fatalf("FormatCurrentWeather() = %q, want %q", got, want)
	}

	if len(requests) != 3 {
		t.Fatalf("FormatCurrentWeather() made %d requests, want 3", len(requests))
	}

	for index, city := range []string{"Astrakhan", "Montreal", "Seattle"} {
		req := requests[index]
		if req.Method != http.MethodGet {
			t.Fatalf("request %d method = %q, want %q", index, req.Method, http.MethodGet)
		}
		if gotCity := req.URL.Query().Get("q"); gotCity != city {
			t.Fatalf("request %d q = %q, want %q", index, gotCity, city)
		}
		if gotKey := req.URL.Query().Get("key"); gotKey != "test-key" {
			t.Fatalf("request %d key = %q, want %q", index, gotKey, "test-key")
		}
		if gotAQI := req.URL.Query().Get("aqi"); gotAQI != "no" {
			t.Fatalf("request %d aqi = %q, want %q", index, gotAQI, "no")
		}
	}
}

func TestFormatCurrentWeatherReturnsErrorOnTransportFailure(t *testing.T) {
	service := newTestService(t, stubHTTPClient{
		do: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("boom")
		},
	})

	if _, err := service.FormatCurrentWeather(context.Background()); err == nil {
		t.Fatal("FormatCurrentWeather() error = nil, want non-nil")
	}
}

func TestFormatCurrentWeatherReturnsErrorOnUnexpectedStatus(t *testing.T) {
	service := newTestService(t, stubHTTPClient{
		do: func(req *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusBadGateway, `{"error":"upstream failure"}`), nil
		},
	})

	if _, err := service.FormatCurrentWeather(context.Background()); err == nil {
		t.Fatal("FormatCurrentWeather() error = nil, want non-nil")
	}
}

func TestFormatCurrentWeatherReturnsErrorOnInvalidJSON(t *testing.T) {
	service := newTestService(t, stubHTTPClient{
		do: func(req *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusOK, `{`), nil
		},
	})

	if _, err := service.FormatCurrentWeather(context.Background()); err == nil {
		t.Fatal("FormatCurrentWeather() error = nil, want non-nil")
	}
}

func TestFormatCurrentWeatherReturnsErrorOnIncompleteResponse(t *testing.T) {
	service := newTestService(t, stubHTTPClient{
		do: func(req *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusOK, `{"location":{"name":"Astrakhan"},"current":{"temp_c":1.2}}`), nil
		},
	})

	if _, err := service.FormatCurrentWeather(context.Background()); err == nil {
		t.Fatal("FormatCurrentWeather() error = nil, want non-nil")
	}
}

func TestFetchRetriesOnTransientError(t *testing.T) {
	calls := 0
	service := newTestService(t, stubHTTPClient{
		do: func(req *http.Request) (*http.Response, error) {
			calls++
			if calls <= 2 {
				return nil, errors.New("network blip")
			}
			return jsonResponse(http.StatusOK, weatherPayload("Astrakhan", 5.0, 3.0, "Clear")), nil
		},
	})

	_, err := service.FormatCurrentWeather(context.Background())
	if err != nil {
		t.Fatalf("FormatCurrentWeather() error = %v, want nil", err)
	}
}

func TestFetchDoesNotRetryOn4xx(t *testing.T) {
	calls := 0
	service := newTestService(t, stubHTTPClient{
		do: func(req *http.Request) (*http.Response, error) {
			calls++
			return jsonResponse(http.StatusUnauthorized, `{"error":"bad key"}`), nil
		},
	})

	_, err := service.FormatCurrentWeather(context.Background())
	if err == nil {
		t.Fatal("FormatCurrentWeather() error = nil, want non-nil")
	}
	if calls != 1 {
		t.Fatalf("expected 1 call for 4xx error, got %d", calls)
	}
}

func newTestService(t *testing.T, client HTTPClient) *Service {
	t.Helper()

	service, err := NewService(client, "test-key")
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}

	service.retryBaseDelay = 0

	return service
}

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func weatherPayload(name string, tempC, feelsLikeC float64, condition string) string {
	return fmt.Sprintf(
		`{"location":{"name":%q},"current":{"temp_c":%.1f,"feelslike_c":%.1f,"condition":{"text":%q}}}`,
		name,
		tempC,
		feelsLikeC,
		condition,
	)
}
