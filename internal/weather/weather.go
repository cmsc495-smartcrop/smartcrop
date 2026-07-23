// Package weather fetches forecast data for a field station's location from
// the National Weather Service (weather.gov) API.
package weather

import (
	"context"
	"fmt"
	"time"

	"github.com/comsma/weatherdotgov/pkg/weathergov"
)

// userAgent is sent on every request, as required by the weather.gov API
// usage policy so the NWS can reach us about our traffic.
const userAgent = "smartcrop, hello@comsma.com"

type Weather struct {
	client *weathergov.ClientWithResponses
}

func NewClient() (*Weather, error) {
	client, err := weathergov.New(userAgent)
	if err != nil {
		return nil, fmt.Errorf("weather: new client: %w", err)
	}

	return &Weather{
		client: client,
	}, nil
}

// ForecastPeriod is a single period of a location's forecast, e.g. "Tuesday"
// or "Tuesday Night".
type ForecastPeriod struct {
	Name                       string
	StartTime                  time.Time
	EndTime                    time.Time
	IsDaytime                  bool
	Temperature                int
	TemperatureUnit            string
	ProbabilityOfPrecipitation int
	ShortForecast              string
	DetailedForecast           string
}

// GetForecastForLocation returns the multi-day forecast for the grid square
// containing the given latitude/longitude.
func (w *Weather) GetForecastForLocation(ctx context.Context, latitude, longitude float64) ([]ForecastPeriod, error) {
	point, err := w.client.PointWithResponse(ctx, float32(latitude), float32(longitude))
	if err != nil {
		return nil, fmt.Errorf("weather: get point: %w", err)
	}
	if point.ApplicationgeoJSON200 == nil {
		return nil, fmt.Errorf("weather: get point: %s", responseError(point.ApplicationproblemJSONDefault, point.StatusCode()))
	}

	props := point.ApplicationgeoJSON200.Properties
	if props.GridId == nil || props.GridX == nil || props.GridY == nil {
		return nil, fmt.Errorf("weather: get point: response missing grid identifiers")
	}

	forecast, err := w.client.GridpointForecastWithResponse(ctx, *props.GridId, *props.GridX, *props.GridY, nil)
	if err != nil {
		return nil, fmt.Errorf("weather: get forecast: %w", err)
	}
	if forecast.ApplicationgeoJSON200 == nil {
		return nil, fmt.Errorf("weather: get forecast: %s", responseError(forecast.ApplicationproblemJSONDefault, forecast.StatusCode()))
	}

	periods := forecast.ApplicationgeoJSON200.Properties.Periods
	if periods == nil {
		return nil, nil
	}

	result := make([]ForecastPeriod, 0, len(*periods))
	for _, p := range *periods {
		period := ForecastPeriod{
			Name:             stringValue(p.Name),
			IsDaytime:        boolValue(p.IsDaytime),
			ShortForecast:    stringValue(p.ShortForecast),
			DetailedForecast: stringValue(p.DetailedForecast),
		}
		if p.TemperatureUnit != nil {
			period.TemperatureUnit = string(*p.TemperatureUnit)
		}
		if p.StartTime != nil {
			period.StartTime = *p.StartTime
		}
		if p.EndTime != nil {
			period.EndTime = *p.EndTime
		}
		if p.Temperature != nil {
			if temp, err := p.Temperature.AsGridpoint12hForecastPeriodTemperature1(); err == nil {
				period.Temperature = temp
			}
		}
		if p.ProbabilityOfPrecipitation != nil && p.ProbabilityOfPrecipitation.Value != nil {
			period.ProbabilityOfPrecipitation = int(*p.ProbabilityOfPrecipitation.Value)
		}
		result = append(result, period)
	}

	return result, nil
}

func responseError(problem *weathergov.Error, statusCode int) string {
	if problem != nil {
		return problem.Detail
	}
	return fmt.Sprintf("unexpected response (status %d)", statusCode)
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func boolValue(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}
