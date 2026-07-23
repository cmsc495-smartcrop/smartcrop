// Package watering derives a watering recommendation for a field station
// from its most recent soil moisture / temperature readings and, when
// available, a short-range precipitation forecast. It has no knowledge of
// the database or HTTP layers — callers translate into and out of its
// input/output types.
package watering

import (
	"fmt"
	"time"
)

const (
	// DrySoilMoisturePercent is the soil moisture level (%) below which the
	// crop is considered too dry and watering is warranted (absent a rain
	// override).
	DrySoilMoisturePercent = 30.0

	// AdequateSoilMoisturePercent is the soil moisture level (%) at or above
	// which no watering is needed regardless of temperature or forecast.
	AdequateSoilMoisturePercent = 50.0

	// HotTemperatureFahrenheit is the current-temperature threshold (°F)
	// that tips a borderline soil-moisture reading toward "water".
	HotTemperatureFahrenheit = 85.0

	// HighRainProbabilityPercent is the chance-of-precipitation (%) within
	// RainOverrideWindow that overrides a "water" verdict for dry soil.
	HighRainProbabilityPercent = 50
)

// RainOverrideWindow is how far into the forecast to look for a
// high-probability rain period when deciding on the rain override.
const RainOverrideWindow = 24 * time.Hour

// Verdict is the recommendation outcome.
type Verdict string

const (
	VerdictWater            Verdict = "water"
	VerdictDoNotWater       Verdict = "do_not_water"
	VerdictInsufficientData Verdict = "insufficient_data"
)

// Reading is a minimal, package-local stand-in for a sensor reading so this
// package doesn't need to import internal/database types.
type Reading struct {
	Value      float64
	RecordedAt time.Time
}

// ForecastPeriod is a minimal, package-local stand-in for a forecast period
// (mirrors the fields of weather.ForecastPeriod actually needed here).
type ForecastPeriod struct {
	StartTime                  time.Time
	ProbabilityOfPrecipitation int
}

// Input bundles everything the recommendation function needs.
type Input struct {
	SoilMoisture *Reading         // nil if no soil_moisture reading exists yet
	Temperature  *Reading         // nil if no temperature reading exists yet
	Forecast     []ForecastPeriod // nil/empty if forecast is unavailable
	Now          time.Time        // injected for deterministic tests; callers pass time.Now()
}

// Recommendation is the output: a verdict plus a short, human-readable
// one-line explanation of which factor(s) drove it.
type Recommendation struct {
	Verdict Verdict
	Reason  string
}

// Recommend applies the package's threshold rules to produce a watering
// recommendation. It is a pure function: no I/O, no randomness (aside from
// the caller-supplied Now), safe to unit test exhaustively.
func Recommend(in Input) Recommendation {
	if in.SoilMoisture == nil {
		return Recommendation{
			Verdict: VerdictInsufficientData,
			Reason:  "No soil moisture reading is available yet for this station.",
		}
	}

	moisture := in.SoilMoisture.Value

	if moisture >= AdequateSoilMoisturePercent {
		return Recommendation{
			Verdict: VerdictDoNotWater,
			Reason:  fmt.Sprintf("Soil moisture is adequate (%.0f%%).", moisture),
		}
	}

	windowHours := int(RainOverrideWindow.Hours())

	if moisture < DrySoilMoisturePercent {
		rainSoon, prob, forecastAvailable := highRainWithin(in.Forecast, in.Now, RainOverrideWindow)
		if rainSoon {
			return Recommendation{
				Verdict: VerdictDoNotWater,
				Reason: fmt.Sprintf(
					"Soil moisture is low (%.0f%%), but rain is expected soon (%d%% chance in the next %dh), so watering can wait.",
					moisture, prob, windowHours,
				),
			}
		}
		if forecastAvailable {
			return Recommendation{
				Verdict: VerdictWater,
				Reason: fmt.Sprintf(
					"Soil moisture is low (%.0f%%) and no significant rain is expected in the next %dh.",
					moisture, windowHours,
				),
			}
		}
		return Recommendation{
			Verdict: VerdictWater,
			Reason: fmt.Sprintf(
				"Soil moisture is low (%.0f%%). Forecast data wasn't available, so this recommendation doesn't account for expected rain.",
				moisture,
			),
		}
	}

	// Borderline band: DrySoilMoisturePercent <= moisture < AdequateSoilMoisturePercent.
	if in.Temperature == nil {
		return Recommendation{
			Verdict: VerdictDoNotWater,
			Reason: fmt.Sprintf(
				"Soil moisture is moderate (%.0f%%) and no temperature reading is available to assess heat stress.",
				moisture,
			),
		}
	}
	if in.Temperature.Value >= HotTemperatureFahrenheit {
		return Recommendation{
			Verdict: VerdictWater,
			Reason: fmt.Sprintf(
				"Soil moisture is moderate (%.0f%%) and it's hot (%.0f°F), so watering is recommended.",
				moisture, in.Temperature.Value,
			),
		}
	}
	return Recommendation{
		Verdict: VerdictDoNotWater,
		Reason: fmt.Sprintf(
			"Soil moisture is moderate (%.0f%%) and temperature is not high enough (%.0f°F) to warrant watering.",
			moisture, in.Temperature.Value,
		),
	}
}

// highRainWithin reports whether any forecast period starting within window
// of now has a precipitation probability at/above HighRainProbabilityPercent.
// The second return value is that period's (highest) probability, for the
// reason string. The third reports whether any forecast data was supplied at
// all, distinguishing "no rain expected" from "no forecast available".
func highRainWithin(periods []ForecastPeriod, now time.Time, window time.Duration) (rainSoon bool, prob int, available bool) {
	if len(periods) == 0 {
		return false, 0, false
	}
	cutoff := now.Add(window)
	for _, p := range periods {
		if p.StartTime.After(cutoff) {
			continue
		}
		if p.ProbabilityOfPrecipitation >= HighRainProbabilityPercent && p.ProbabilityOfPrecipitation > prob {
			rainSoon = true
			prob = p.ProbabilityOfPrecipitation
		}
	}
	return rainSoon, prob, true
}
