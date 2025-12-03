package pricing

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vantutran2k1/flowfleet/internal/core/domain"
)

func TestStandardStrategy_CalculatePrice(t *testing.T) {
	strategy := NewStandardStrategy()

	tests := []struct {
		name     string
		input    domain.PricingInput
		expected int
		wantErr  bool
	}{
		{
			name: "Bike Trip",
			input: domain.PricingInput{
				DistanceMeters: 10000,
				Vehicle:        domain.VehicleBike,
			},
			expected: 1000,
			wantErr:  false,
		},
		{
			name: "Van Trip",
			input: domain.PricingInput{
				DistanceMeters: 5000,
				Vehicle:        domain.VehicleVan,
			},
			expected: 2000,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := strategy.CalculatePrice(context.Background(), tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}
