package impact

import (
	"math"
	"testing"

	"github.com/bmizerany/assert"
)

var (
	mockShort = []float64{0.2, 0, 0.4, 0, 0.1, 0.5, 0.2, 0.4, 0, 0, 0.1, 0.6, 0.1, 0.3, 0.1, 0.1, 0.2, 0.3, 0.1, 0.1}
)

func TestComparisons(t *testing.T) {
	greaterCount := gt(0.3, mockShort)
	assert.Equal(t, 4, greaterCount)

	lessCount := lt(0.3, mockShort)
	assert.Equal(t, 14, lessCount)
}

func TestDetect(t *testing.T) {
	// get the pvalue and operator for the test
	p, op, err := DetectImpact(mockShort[0:14], mockShort[14:20], 1000)
	assert.Equal(t, nil, err)
	assert.Tf(t, p > 0.1 && p < 0.25, "pvalue for Detect should be within [0.1, 0.2]")
	assert.Equalf(t, LESS_THAN, op, "the series detection should show a decrease")
}

func equal(a, b float64) bool {
	eps := math.Abs(a - math.Nextafter(a, 1))
	abs := math.Abs(b - a)

	return abs <= eps
}
