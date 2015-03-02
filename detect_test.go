package impact

import (
	"testing"

	"github.com/bmizerany/assert"
)

var (
	mockData = []float64{0.2, 0, 0.4, 0, 0.1, 0.5, 0.2, 0.4, 0, 0, 0.1, 0.6, 0.1, 0.3, 0.1, 0.1, 0.2, 0.3, 0.1, 0.1}
	detector = NewDetector()
)

func TestSmooth(t *testing.T) {
	smoothed := detector.smooth(mockData)

	// check the first couple of smoothed values
	assert.Equalf(t, 0.125, smoothed[0], "incorrect smoothed value")
	assert.Equalf(t, 0.125, smoothed[1], "incorrect smoothed value")
	assert.Equalf(t, 0.0, smoothed[2], "incorrect smoothed value")
	assert.Equalf(t, 0.125, smoothed[3], "incorrect smoothed value")
	assert.Equalf(t, 0.25, smoothed[4], "incorrect smoothed value")
}

func TestDiff(t *testing.T) {
	mockDiff := diff(mockData)

	// make sure the difference is the right length
	assert.Equal(t, len(mockData)-1, len(mockDiff))

	// make sure the first couple differences are correct
	assert.Equalf(t, -0.2, mockDiff[0], "incorrect vector difference")
	assert.Equalf(t, 0.4, mockDiff[1], "incorrect vector difference")
}

func TestComparisons(t *testing.T) {
	greaterCount := gt(0.3, mockData)
	assert.Equal(t, 4, greaterCount)

	lessCount := lt(0.3, mockData)
	assert.Equal(t, 14, lessCount)
}

func TestWalk(t *testing.T) {
	n := 10
	randomWalk := detector.walk(10.0, n, mockData)

	assert.Equal(t, n, len(randomWalk))

	randomMean := mean(randomWalk)
	assert.Tf(t, randomMean > 7.0, "mean of random walk is too low")
	assert.Tf(t, randomMean < 13.0, "mean of random walk is too high")
}

func TestDetect(t *testing.T) {
	// get the pvalue and operator for the test
	p, op := detector.Detect(mockData[0:14], mockData[14:20], 200)
	assert.Tf(t, p < 0.1, "pvalue for Detect should be small (likely < 0.05)")
	assert.Equalf(t, LESS_THAN, op, "the series detection should show a decrease")
}
