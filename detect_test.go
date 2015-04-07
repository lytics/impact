package impact

import (
	"math"
	"testing"

	"github.com/bmizerany/assert"
)

var (
	mockShort = []float64{0.2, 0, 0.4, 0, 0.1, 0.5, 0.2, 0.4, 0, 0, 0.1, 0.6, 0.1, 0.3, 0.1, 0.1, 0.2, 0.3, 0.1, 0.1}
)

func TestSmooth(t *testing.T) {
	smoothed := smooth(mockShort)

	// check the first couple of smoothed values
	assert.Equalf(t, true, equal(0.2, smoothed[0]), "incorrect smoothed value")
	assert.Equalf(t, true, equal(0.15, smoothed[1]), "incorrect smoothed value")
	assert.Equalf(t, true, equal(0.14, smoothed[2]), "incorrect smoothed value")
	assert.Equalf(t, true, equal(0.2, smoothed[3]), "incorrect smoothed value")
	assert.Equalf(t, true, equal(0.24, smoothed[4]), "incorrect smoothed value")
}

func TestDiff(t *testing.T) {
	mockDiff := diff(mockShort)

	// make sure the difference is the right length
	assert.Equal(t, len(mockShort)-1, len(mockDiff))

	// make sure the first couple differences are correct
	assert.Equalf(t, true, equal(-0.2, mockDiff[0]), "incorrect vector difference")
	assert.Equalf(t, true, equal(0.4, mockDiff[1]), "incorrect vector difference")
}

func TestComparisons(t *testing.T) {
	greaterCount := gt(0.3, mockShort)
	assert.Equal(t, 4, greaterCount)

	lessCount := lt(0.3, mockShort)
	assert.Equal(t, 14, lessCount)
}

func TestWalk(t *testing.T) {
	n := 10
	randomWalk := walk(10.0, n, mockShort)

	assert.Equal(t, n, len(randomWalk))

	randomMean := mean(randomWalk)
	assert.Tf(t, randomMean > 7.0, "mean of random walk is too low")
	assert.Tf(t, randomMean < 13.0, "mean of random walk is too high")
}

func TestDetect(t *testing.T) {
	// get the pvalue and operator for the test
	p, op := DetectImpact(mockShort[0:14], mockShort[14:20], 200)
	assert.Tf(t, p < 0.1, "pvalue for Detect should be small (likely < 0.05)")
	assert.Equalf(t, LESS_THAN, op, "the series detection should show a decrease")
}

func equal(a, b float64) bool {
	eps := math.Abs(a - math.Nextafter(a, 1))
	abs := math.Abs(b - a)

	return abs <= eps
}
