package impact

import (
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/drewlanenga/govector"
)

// the operator indicates whether the candidate series has increased,
// decreased or stayed largely the same
type Operator int

type series govector.Vector

const (
	EQUALS       Operator = 0
	GREATER_THAN Operator = 1
	LESS_THAN    Operator = 2
)

var (
	smoother uint = 2 // the amount of smoothing on either side
	rnd           = rand.New(rand.NewSource(time.Now().UnixNano()))
	rndMutex      = &sync.Mutex{}
)

func walks(niter, nsteps, ncpu int, start float64, history govector.Vector) series {
	destinations := make(series, niter)

	steps := history.Diff()

	c := make(chan int, ncpu)
	for i := 0; i < niter; i++ {
		go destinations.walk(i, nsteps, start, steps, c)
	}

	// drain the channel
	for i := 0; i < ncpu; i++ {
		<-c // wait for one task to complete
	}

	// all done
	return destinations
}

// take random steps in a walk based on the `diff`.  (`diff` is a bunch of steps.)
func (s series) walk(i, nsteps int, start float64, diff govector.Vector, c chan int) {
	walkrnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	n := len(diff)
	dest := start
	for i := 0; i < nsteps; i++ {
		which := walkrnd.Intn(n)
		dest += diff[which]
	}

	s[i] = dest
	c <- 1 // signal that the walk has finished
}

// DetectImpact performs Monte Carlo based changepoint detection between two disjoint
// and adjacent subseries of a larger time series.  Increase `niter` to improve
// accuracy of the detection.
func DetectImpact(x1, x2 []float64, niter int) (float64, Operator, error) {
	v1, err := govector.AsVector(x1)
	if err != nil {
		return 0.0, EQUALS, err
	}

	v2, err := govector.AsVector(x2)
	if err != nil {
		return 0.0, EQUALS, err
	}

	x1smooth := v1.Smooth(smoother, smoother)
	x2smooth := v2.Smooth(smoother, smoother)

	x1diff := x1smooth.Diff()

	ncpu, _ := strconv.Atoi(os.Getenv("GOMAXPROCS"))
	if ncpu == 0 {
		ncpu = runtime.NumCPU()
	}
	runtime.GOMAXPROCS(ncpu)

	// the final destinations of a bunch of random walks
	simDest := walks(niter, len(x2), ncpu, x1smooth[len(x1)-1], x1diff)

	realDest := x2smooth[len(x2)-1]

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

	return p, op, nil
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
