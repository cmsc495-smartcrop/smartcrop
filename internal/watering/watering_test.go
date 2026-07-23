package watering

import (
	"strings"
	"testing"
	"time"
)

func TestRecommend(t *testing.T) {
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		in             Input
		wantVerdict    Verdict
		reasonContains string
	}{
		{
			name: "dry and no rain in forecast recommends water",
			in: Input{
				SoilMoisture: &Reading{Value: 22},
				Forecast:     []ForecastPeriod{{StartTime: now.Add(2 * time.Hour), ProbabilityOfPrecipitation: 10}},
				Now:          now,
			},
			wantVerdict:    VerdictWater,
			reasonContains: "no significant rain",
		},
		{
			name: "dry with high rain chance within 24h holds off",
			in: Input{
				SoilMoisture: &Reading{Value: 22},
				Forecast:     []ForecastPeriod{{StartTime: now.Add(2 * time.Hour), ProbabilityOfPrecipitation: 60}},
				Now:          now,
			},
			wantVerdict:    VerdictDoNotWater,
			reasonContains: "rain",
		},
		{
			name: "dry with high rain chance beyond 24h still recommends water",
			in: Input{
				SoilMoisture: &Reading{Value: 22},
				Forecast:     []ForecastPeriod{{StartTime: now.Add(30 * time.Hour), ProbabilityOfPrecipitation: 90}},
				Now:          now,
			},
			wantVerdict:    VerdictWater,
			reasonContains: "no significant rain",
		},
		{
			name: "adequate moisture never recommends watering",
			in: Input{
				SoilMoisture: &Reading{Value: 65},
				Temperature:  &Reading{Value: 95},
				Forecast:     []ForecastPeriod{{StartTime: now.Add(2 * time.Hour), ProbabilityOfPrecipitation: 0}},
				Now:          now,
			},
			wantVerdict:    VerdictDoNotWater,
			reasonContains: "adequate",
		},
		{
			name: "borderline and hot recommends water",
			in: Input{
				SoilMoisture: &Reading{Value: 40},
				Temperature:  &Reading{Value: 90},
				Now:          now,
			},
			wantVerdict:    VerdictWater,
			reasonContains: "hot",
		},
		{
			name: "borderline and not hot does not recommend watering",
			in: Input{
				SoilMoisture: &Reading{Value: 40},
				Temperature:  &Reading{Value: 70},
				Now:          now,
			},
			wantVerdict:    VerdictDoNotWater,
			reasonContains: "not high enough",
		},
		{
			name: "borderline with missing temperature does not recommend watering",
			in: Input{
				SoilMoisture: &Reading{Value: 40},
				Now:          now,
			},
			wantVerdict:    VerdictDoNotWater,
			reasonContains: "no temperature reading",
		},
		{
			name: "missing soil moisture is insufficient data",
			in: Input{
				Temperature: &Reading{Value: 95},
				Forecast:    []ForecastPeriod{{StartTime: now.Add(2 * time.Hour), ProbabilityOfPrecipitation: 90}},
				Now:         now,
			},
			wantVerdict:    VerdictInsufficientData,
			reasonContains: "No soil moisture reading",
		},
		{
			name: "dry with no forecast data recommends water and notes it",
			in: Input{
				SoilMoisture: &Reading{Value: 22},
				Now:          now,
			},
			wantVerdict:    VerdictWater,
			reasonContains: "Forecast data wasn't available",
		},
		{
			name: "moisture exactly at dry threshold is borderline not dry",
			in: Input{
				SoilMoisture: &Reading{Value: DrySoilMoisturePercent},
				Temperature:  &Reading{Value: 70},
				Now:          now,
			},
			wantVerdict:    VerdictDoNotWater,
			reasonContains: "moderate",
		},
		{
			name: "moisture exactly at adequate threshold is adequate not borderline",
			in: Input{
				SoilMoisture: &Reading{Value: AdequateSoilMoisturePercent},
				Now:          now,
			},
			wantVerdict:    VerdictDoNotWater,
			reasonContains: "adequate",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Recommend(tc.in)
			if got.Verdict != tc.wantVerdict {
				t.Errorf("Verdict = %q, want %q (reason: %q)", got.Verdict, tc.wantVerdict, got.Reason)
			}
			if !strings.Contains(got.Reason, tc.reasonContains) {
				t.Errorf("Reason = %q, want it to contain %q", got.Reason, tc.reasonContains)
			}
		})
	}
}
