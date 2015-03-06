package impact

import (
	"fmt"
	"math"
	"sort"

	"github.com/bobhancock/gomatrix/matrix"
)

type splitter struct {
	Index  int     // location of the changepoint
	Energy float64 // "energy" released when cluster splits
}

func newSplitter() *splitter {
	return &splitter{-1, math.Inf(-1)}
}

func newSplitters(n int) []*splitter {
	splitters := make([]*splitter, n)
	for i := 0; i < n; i++ {
		splitters[i] = newSplitter()
	}
	return splitters
}

type splitSummary struct {
	Changes []int
	Best    float64
}

type permSummary struct {
	P float64
	R int
}

// DetectChanges implements divisive changepoint detection to identify
// structural changes to x.  `sig` determines significance level, `R` determines
// number of permutations to run during permutation testing, and `minSize`
// determines the minimum size of the series to detect.
func DetectChanges(x []float64, sig float64, R int, minSize int) ([]int, error) {
	if sig < 0 || sig > 1.0 {
		return nil, fmt.Errorf("sig (%v) should be bound [0, 1]", sig)
	}

	if minSize < 2 {
		return nil, fmt.Errorf("minSize (%d) must be greater than 1", minSize)
	}

	n := len(x)

	// initialize k to changepoints
	k := n

	// assume changes occur at beginning and end of series
	changes := []int{0, n}
	splitters := newSplitters(n)

	distance := vectorDistance(x)

	for k > 0 {
		split := eSplit(changes, distance, minSize, false, splitters)
		newestChangePoint := split.Changes[len(split.Changes)-1]

		// not able to meet minimum size constraint
		if newestChangePoint == -1 {
			break
		}

		result := sigTest(distance, R, changes, minSize, split.Best, splitters)

		// change point not significant
		if result.P > sig {
			break
		}

		// update set of change points
		changes = split.Changes
		k--
	}

	// remove the last (insignificant) changepoint
	significant := changes[0 : len(changes)-1]

	// sort them in sequential order (ordered natively by discovery)
	sort.Sort(sort.IntSlice(significant))

	return significant, nil
}

func eSplit(changes []int, distance *matrix.DenseMatrix, minSize int, forSim bool, splitters []*splitter) splitSummary {
	// copy changes into splits
	splits := copyAndSort(changes)

	best := newSplitter()

	ii := -1

	// if the procedure is being used for a significance test
	if forSim {
		for i := 1; i < len(splits); i++ {
			split := splitPoint(splits[i-1], splits[i]-1, distance, minSize)
			if split.Energy > best.Energy {
				ii = splits[i-1]

				best = split // best split found so far
			}
		}

		changes = append(changes, best.Index)
		return splitSummary{changes, best.Energy}
	}

	for i := 1; i < len(splits); i++ {
		isplitter := splitters[splits[i-1]]
		if isplitter.Index == -1 {
			isplitter = splitPoint(splits[i-1], splits[i]-1, distance, minSize)
		}

		if isplitter.Energy > best.Energy {
			ii = splits[i-1]
			best = isplitter
		}
	}

	changes = append(changes, best.Index)
	splitters[ii].Index = 0    // update to account for newly proposed changepoint
	splitters[ii].Energy = 0.0 // update to account for newly proposed changepoint

	return splitSummary{changes, best.Energy}
}

// this implementation is of complexity O(n^2) to find each change point
// so if k change points are found, complexity is O(kn^2)
func splitPoint(start int, end int, distance *matrix.DenseMatrix, minSize int) *splitter {
	// interval is too small to split
	if (end - start + 1) < 2*minSize {
		return newSplitter()
	}
	best := newSplitter()

	dist := distance.Copy()

	// now represents number of data points
	end = end - start + 1

	tau1 := minSize
	tau2 := minSize << 1

	cut1 := subsetMatrix(dist, numericRange(0, tau1-1), numericRange(0, tau1-1))
	cut2 := subsetMatrix(dist, numericRange(tau1, tau2-1), numericRange(tau1, tau2-1))
	cut3 := subsetMatrix(dist, numericRange(0, tau1-1), numericRange(tau1, tau2-1))

	// within distance for left cluster
	a := matrixSum(cut1) / 2.0

	// within distance for right cluster
	b1 := matrixSum(cut2) / 2.0

	// between distance for both clusters
	ab1 := matrixSum(cut3)

	energy := calculateEnergy(a, b1, ab1, tau1, tau2)
	if energy > best.Energy {
		best.Index = tau1 + start
		best.Energy = energy
	}

	// shift right cluster
	tau2 += 1

	b := initVector(end+1, b1)
	ab := initVector(end+1, ab1)

	for tau2 <= end {
		b[tau2] = b[tau2-1] + matrixSum(subsetMatrix(distance, numericRange(tau2-1, tau2-1), numericRange(tau1, tau2-2)))
		ab[tau2] = ab[tau2-1] + matrixSum(subsetMatrix(distance, numericRange(tau2-1, tau2-1), numericRange(0, tau1-1)))

		energy = calculateEnergy(a, b[tau2], ab[tau2], tau1, tau2)
		if energy > best.Energy {
			best.Index = tau1 + start
			best.Energy = energy
		}
		tau2++
	}

	// shift left cluster
	tau1 += 1

	for {
		tau2 = tau1 + minSize
		if tau2 > end {
			break
		}

		addA := matrixSum(subsetMatrix(distance, numericRange(tau1-1, tau1-1), numericRange(0, tau1-2)))
		addB := matrixSum(subsetMatrix(distance, numericRange(tau1-1, tau1-1), numericRange(tau1, tau2-2)))

		// update within distance for left cluster
		a += addA

		// iterate over possible endings for right cluster (tau2)
		for tau2 <= end {
			// update within distance for right cluster
			addB += distance.Get(tau1-1, tau2-1)
			b[tau2] -= addB

			// update between cluster distance
			ab[tau2] += addB - addA
			energy = calculateEnergy(a, b[tau2], ab[tau2], tau1, tau2)
			if energy > best.Energy {
				best.Index = tau1 + start
				best.Energy = energy
			}
			tau2++
		}
		tau1++
	}
	return best
}

func calculateEnergy(a, b, ab float64, tau1, tau2 int) float64 {
	info := scaleAB(ab, tau1, tau2) - scaleB(b, tau1, tau2) - scaleA(a, tau1, tau2)
	tau := float64(tau1*(tau2-tau1)) / float64(tau2)
	return info * tau
}

func scaleA(a float64, tau1, tau2 int) float64 {
	return 2.0 * a / float64(tau1*(tau1-1))
}

func scaleB(b float64, tau1, tau2 int) float64 {
	return 2.0 * b / float64((tau2-tau1-1)*(tau2-tau1))
}

func scaleAB(ab float64, tau1, tau2 int) float64 {
	return 2.0 * ab / float64((tau2-tau1)*tau1)
}

// create a numeric slice with range [start, end] inclusively
func numericRange(start, end int) []int {
	n := end - start + 1
	ints := make([]int, n)
	for i := 0; i < n; i++ {
		ints[i] = i + start
	}
	return ints
}

// return a subset of m given vectors of row and column indeces
// TODO: we might be able to get away with a subset of the row/col for the single col/row extraction
func subsetMatrix(m *matrix.DenseMatrix, rows []int, cols []int) *matrix.DenseMatrix {
	elements := make([]float64, len(rows)*len(cols))
	subset := matrix.MakeDenseMatrix(elements, len(rows), len(cols))

	for newRowIndex, oldRowIndex := range rows {
		for newColIndex, oldColIndex := range cols {
			subset.Set(newRowIndex, newColIndex, m.Get(oldRowIndex, oldColIndex))
		}
	}

	return subset
}

// initialize vector of length n to a specific value
func initVector(n int, value float64) []float64 {
	x := make([]float64, n)
	for i := 0; i < n; i++ {
		x[i] = value
	}
	return x
}

// distance between each point
func vectorDistance(x []float64) *matrix.DenseMatrix {
	elements := make([]float64, 0, len(x)*len(x))
	for _, xi := range x {
		xdist := make([]float64, len(x))
		for j, xj := range x {
			xdist[j] = math.Abs(xi - xj)
		}
		elements = append(elements, xdist...)
	}

	return matrix.MakeDenseMatrix(elements, len(x), len(x))
}

func matrixSum(m *matrix.DenseMatrix) float64 {
	sum := 0.0
	for i := 0; i < m.Cols(); i++ {
		sum += m.SumCol(i)
	}
	return sum
}

func sigTest(distance *matrix.DenseMatrix, R int, changes []int, minSize int, obs float64, splitters []*splitter) permSummary {
	if R == 0 {
		return permSummary{0.0, -1}
	}

	over := 0
	for f := 0; f < R; f++ {
		D1 := permCluster(distance, changes)
		split := eSplit(changes, D1, minSize, true, splitters)
		if split.Best > obs {
			over++
		}
	}

	// pad the pvalue by 1 success
	p := float64(over+1) / float64(R+1)
	return permSummary{p, R}
}

func permCluster(d *matrix.DenseMatrix, changes []int) *matrix.DenseMatrix {
	points := copyAndSort(changes)

	for i := 0; i < len(points)-1; i++ { // number of clusters
		index := numericRange(points[i], points[i+1]-1) // shuffle within clusters by permuting matrix columns and rows
		u := shuffle(index)
		for ii, ui := range u {
			d.Set(ii, ii, d.Get(ui, ui))
		}
	}
	return d
}

// maybe not a good copy
func copyAndSort(x []int) []int {
	y := make([]int, len(x))
	for i, xi := range x {
		y[i] = xi
	}
	// sort the current set of change points
	sort.Sort(sort.IntSlice(y))

	return y
}

func shuffle(x []int) []int {
	// make it safe
	rndMutex.Lock()
	index := rnd.Perm(len(x))
	rndMutex.Unlock()

	y := make([]int, len(x))
	for xindex, yindex := range index {
		y[yindex] = x[xindex]
	}

	return y
}
