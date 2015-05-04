package impact

import (
	"fmt"
)

func Example() {
	niter := 1000
	series := []float64{0.2, 0, 0.4, 0, 0.1, 0.5, 0.2, 0.4, 0, 0, 0.1, 0.6, 0.1, 0.3, 0.1, 0.1, 0.2, 0.3, 0.1, 0.1}

	// detect changepoints
	significance := 0.05
	minSize := 3
	changes, _ := DetectChanges(series, significance, niter, minSize)
	fmt.Println(changes)
	// Output: [0 8 20]

	// detect impact
	_, _, _ = DetectImpact(series[changes[0]:changes[1]], series[changes[1]:changes[2]], niter)
}
