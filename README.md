Impact
======

[![Build Status](https://travis-ci.org/lytics/impact.svg?branch=master)](https://travis-ci.org/lytics/impact) [![GoDoc](https://godoc.org/github.com/lytics/impact?status.svg)](https://godoc.org/github.com/lytics/impact)

Lightweight bootstrap testing for detecting causal impact to timeseries in [Go](https://golang.org).

## Purpose

**Impact detects significant changes to the location of a time series**.  For a candidate point in time for change detection, it uses the structure of the preceding data to determine the probability of the subsequent data arriving to its final location.  A low probability indicates a significant departure in location, or a *casual impact* to the series.

Because of the nature of the underlying Monte Carlo simulation, Impact is free of any distributional assumptions, and fit for use in any processes whose likelihood function is either unknown or dynamic &mdash; **it is location, scale and distribution free**.

## Design

Impact requires a *reference* sub-series, which serves to provide information on the structural profile of the series.  It also requires a *candidate* sub-series which supplies evidence for or against the location change.

Impact assumes that both the *reference* and *candidate* subseries come from the same process under the same generative parameters.  Under this assumption, it performs Monte Carlo simulation (via bootstrap resampling) to simulate possible alternatives to the *candidate* subseries using data from the *reference* subseries.  The destination of each simulated walk is compared against the realized value from the *candidate* subseries, and the percentage of destinations considered as or more extreme (in terms of absolute deviation in location) create the "p-value" for the test.

## Example

Consider the following downward trending process, which is divided into two disjoint series &mdash; the *Reference* series in solid black, and the *Candidate* series in dotted black.

![negativewalk](https://cloud.githubusercontent.com/assets/3698679/6422052/3c21eb06-be89-11e4-889f-f1718207d53a.png)

In order to determine if the start of the *Candidate* series indicates a causal disruption to the sequence, we simulate a large number of alternatives and deem that since the observed *Candidate* series is more extreme (in terms of final destination) than any of the simulations, that the start of the *Candidate* series indicates a causal disruption.

![negativewalk_20](https://cloud.githubusercontent.com/assets/3698679/6422055/3c231cec-be89-11e4-966e-265bcd50766f.png)

Alternatively, consider the following upward trending process and its corresponding *Reference* and *Candidate* sub-series.

![positivewalk](https://cloud.githubusercontent.com/assets/3698679/6422053/3c224d58-be89-11e4-96e7-219acda4691e.png)

We likewise simulate a large number of alternatives.  Since the realized *Candidate* sub-series lies well within the range of simulated alternatives, there's no evidence of a causal disruption at this point in the series.

![positivewalk_20](https://cloud.githubusercontent.com/assets/3698679/6422054/3c22e862-be89-11e4-8513-18a06925f772.png)

*Note that although only 20 simulated alternatives are shown in each figure, that in practice the number of bootstrap resamples should be large enough to yield conclusive results &mdash; definitely upwards of 1,000.

## Usage

```go
package main

import (
	"fmt"

	"github.com/lytics/impact"
)

func main() {
	niter := 1000
	series := []float64{0.2, 0, 0.4, 0, 0.1, 0.5, 0.2, 0.4, 0, 0, 0.1, 0.6, 0.1, 0.3, 0.1, 0.1, 0.2, 0.3, 0.1, 0.1}
	detector := NewDetector()

	// sub-series for detection must be chosen a priori
	p, op := detector.Detect(series[0:14], series[14:20], niter)
	fmt.Println(p, op)

	// Note that because of the nature of bootstrapping, the p-value from the test is subject to minor fluctuations.
	// To get a more accurate/consistent p-value, increase the number of iterations in the detection.
}

```
