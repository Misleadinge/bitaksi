package haversine

import (
	"math"
	"testing"
)

func TestDistance(t *testing.T) {
	tests := []struct {
		name      string
		lat1      float64
		lon1      float64
		lat2      float64
		lon2      float64
		expected  float64
		tolerance float64
	}{
		{
			name:      "Istanbul to Ankara",
			lat1:      41.0082,
			lon1:      28.9784,
			lat2:      39.9334,
			lon2:      32.8597,
			expected:  350.0, // Approximately 350 km
			tolerance: 10.0,
		},
		{
			name:      "Same point",
			lat1:      41.0082,
			lon1:      28.9784,
			lat2:      41.0082,
			lon2:      28.9784,
			expected:  0.0,
			tolerance: 0.1,
		},
		{
			name:      "Short distance",
			lat1:      41.0082,
			lon1:      28.9784,
			lat2:      41.0182,
			lon2:      28.9884,
			expected:  1.4, // Approximately 1.4 km
			tolerance: 0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Distance(tt.lat1, tt.lon1, tt.lat2, tt.lon2)
			diff := math.Abs(result - tt.expected)
			if diff > tt.tolerance {
				t.Errorf("Distance() = %v, expected approximately %v (tolerance: %v)", result, tt.expected, tt.tolerance)
			}
		})
	}
}
