package weather

import (
	"context"
	"encoding/json"
	"math"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	currentWeatherURL = "https://api.weatherapi.com/v1/current.json"
	defaultTimeout    = 5 * time.Second
	maxRetries        = 3
	retryBaseDelay    = 500 * time.Millisecond
)

// HTTPClient is the minimal client contract needed for WeatherAPI requests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Service struct {
	client         HTTPClient
	apiKey         string
	cities         []string
	maxRetries     int
	retryBaseDelay time.Duration
}

func NewService(client HTTPClient, apiKey string) (*Service, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, errors.New("WEATHER_API_KEY is required")
	}

	if client == nil {
		client = &http.Client{Timeout: defaultTimeout}
	}

	return &Service{
		client:         client,
		apiKey:         apiKey,
		cities:         []string{"Astrakhan", "Montreal", "Seattle"},
		maxRetries:     maxRetries,
		retryBaseDelay: retryBaseDelay,
	}, nil
}

func (s *Service) FormatCurrentWeather(ctx context.Context) (string, error) {
	if s == nil {
		return "", errors.New("service is nil")
	}
	if s.client == nil {
		return "", errors.New("http client is nil")
	}
	if strings.TrimSpace(s.apiKey) == "" {
		return "", errors.New("WEATHER_API_KEY is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	lines := make([]string, 0, len(s.cities))
	for _, city := range s.cities {
		line, err := s.fetchCurrentWeather(ctx, city)
		if err != nil {
			return "", err
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n"), nil
}

func (s *Service) fetchCurrentWeather(ctx context.Context, city string) (string, error) {
	var lastErr error
	for attempt := range s.maxRetries {
		if attempt > 0 {
			delay := s.retryBaseDelay * time.Duration(1<<(attempt-1))
			select {
			case <-ctx.Done():
				return "", fmt.Errorf("fetch %s weather: %w", city, ctx.Err())
			case <-time.After(delay):
			}
		}

		result, err := s.doFetch(ctx, city)
		if err == nil {
			return result, nil
		}
		lastErr = err

		if !isRetryable(err) {
			return "", err
		}
	}
	return "", lastErr
}

type statusError struct {
	code int
	msg  string
}

func (e *statusError) Error() string { return e.msg }

func isRetryable(err error) bool {
	var se *statusError
	if errors.As(err, &se) {
		return se.code >= 500
	}
	return true // network errors are retryable
}

func (s *Service) doFetch(ctx context.Context, city string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, currentWeatherURL, nil)
	if err != nil {
		return "", fmt.Errorf("build %s request: %w", city, err)
	}

	query := req.URL.Query()
	query.Set("key", s.apiKey)
	query.Set("q", city)
	query.Set("aqi", "no")
	req.URL.RawQuery = query.Encode()

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch %s weather: %w", city, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := &statusError{
			code: resp.StatusCode,
			msg:  fmt.Sprintf("fetch %s weather: unexpected status %s", city, resp.Status),
		}
		return "", err
	}

	payload, err := decodeCurrentWeather(resp)
	if err != nil {
		return "", fmt.Errorf("decode %s weather: %w", city, err)
	}

	return fmt.Sprintf(
		"%s: %.0f\u00b0 (%.0f\u00b0) / %s",
		payload.name,
		math.Round(payload.tempC),
		math.Round(payload.feelsLikeC),
		payload.conditionText,
	), nil
}

type apiCurrentWeather struct {
	Location struct {
		Name *string `json:"name"`
	} `json:"location"`
	Current struct {
		TempC      *float64 `json:"temp_c"`
		FeelsLikeC *float64 `json:"feelslike_c"`
		Condition  struct {
			Text *string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
}

type currentWeather struct {
	name          string
	tempC         float64
	feelsLikeC    float64
	conditionText string
}

func decodeCurrentWeather(resp *http.Response) (currentWeather, error) {
	var data apiCurrentWeather
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return currentWeather{}, err
	}

	if data.Location.Name == nil ||
		data.Current.TempC == nil ||
		data.Current.FeelsLikeC == nil ||
		data.Current.Condition.Text == nil {
		return currentWeather{}, errors.New("incomplete response")
	}

	return currentWeather{
		name:          *data.Location.Name,
		tempC:         *data.Current.TempC,
		feelsLikeC:    *data.Current.FeelsLikeC,
		conditionText: *data.Current.Condition.Text,
	}, nil
}
