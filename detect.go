package impact

import (
	"math/rand"
	"sync"
	"time"
)

type Detector struct{}

// the operator indicates whether the candidate series has increased,
// decreased or stayed largely the same
type Operator int

const (
	EQUALS       Operator = 0
	GREATER_THAN Operator = 1
	LESS_THAN    Operator = 2
)

var (
	smoother = 2 // the amount of smoothing on either side
	rnd      = rand.New(rand.NewSource(time.Now().UnixNano()))
	rndMutex = &sync.Mutex{}
)

//
//  TODO: Search for off by one index errors since porting from R
//

func NewDetector() *Detector {
	return &Detector{}
}

func (d *Detector) smooth(x []float64) []float64 {
	n := len(x)

	nsmooth := n - smoother

	smoothed := make([]float64, n)

	for i := smoother + 1; i < nsmooth; i++ {
		window := x[(i - smoother):(i + smoother)]
		smoothed[i] = mean(window)
	}

	// pad beginning with first mean
	for i := 0; i < smoother; i++ {
		smoothed[i] = smoothed[smoother+1]
	}

	// pad end with last mean
	for i := n - smoother + 1; i < n; i++ {
		smoothed[i] = smoothed[n-smoother]
	}

	return smoothed
}

// smooth the two series adjacently to borrow information on the boundaries
func (d *Detector) smoothSeries(x1, x2 []float64) ([]float64, []float64) {
	n1 := len(x1)
	n2 := len(x2)

	x1 = append(x1, x2...)
	smoothed := d.smooth(x1)
	return smoothed[0:n1], smoothed[n1:(n1 + n2)]
}

// take random steps in a walk based on the `diff`.  (`diff` is a bunch of steps.)
func (d *Detector) walk(start float64, n int, diff []float64) []float64 {
	simulated := make([]float64, n)

	// where we start our walk, simulate each step
	value := start
	for i := 0; i < n; i++ {
		step := sample(diff)
		value += step
		simulated[i] = value
	}
	return simulated
}

// Perform Monte Carlo based changepoint detection between two disjoint and adjacent subseries of
// a larger time series.  Increase `niter` to improve accuracy of the detection.
func (d *Detector) Detect(x1, x2 []float64, niter int) (float64, Operator) {
	x1smooth, x2smooth := d.smoothSeries(x1, x2)

	n1 := len(x1)
	n2 := len(x2)

	x1diff := diff(x1smooth)

	// the final destinations of a bunch of random walks
	simDest := make([]float64, niter)
	for i := 0; i < niter; i++ {
		walk := d.walk(x1smooth[n1-1], n2, x1diff)
		simDest[i] = walk[n2-1]
	}

	realDest := x2smooth[n2-1]

	plower := float64(lt(realDest, simDest)) / float64(niter)
	pupper := float64(gt(realDest, simDest)) / float64(niter)

	p := 1.0
	op := EQUALS

	if plower < pupper {
		p = plower
		op = LESS_THAN
	} else if pupper < plower {
		p = pupper
		op = GREATER_THAN
	}

	return p, op
}

// count the number of xs greater than x
func gt(x float64, xs []float64) int {
	count := 0
	for _, value := range xs {
		if x < value {
			count++
		}
	}

	return count
}

// count the number of xs less than x
func lt(x float64, xs []float64) int {
	count := 0
	for _, value := range xs {
		if x > value {
			count++
		}
	}

	return count
}

// sample one entry from the vector
func sample(x []float64) float64 {
	rndMutex.Lock()
	defer rndMutex.Unlock()

	index := rnd.Intn(len(x))
	return x[index]
}

// calculate the average of the vector
func mean(x []float64) float64 {
	sum := 0.0
	for _, value := range x {
		sum += value
	}

	return sum / float64(len(x))
}

// calculate a vector of differences
func diff(x []float64) []float64 {
	difference := make([]float64, len(x)-1)

	for i := 0; i < len(x)-1; i++ {
		difference[i] = x[i+1] - x[i]
	}

	return difference
}
