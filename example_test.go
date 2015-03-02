package impact

import (
	"fmt"
)

func Example() {
	niter := 1000
	series := []float64{0.2, 0, 0.4, 0, 0.1, 0.5, 0.2, 0.4, 0, 0, 0.1, 0.6, 0.1, 0.3, 0.1, 0.1, 0.2, 0.3, 0.1, 0.1}
	detector := NewDetector()

	// sub-series for detection must be chosen a priori
	p, op := detector.Detect(series[0:14], series[14:20], niter)
	fmt.Println(p, op)
}
