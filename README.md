Changepoint
===========

[![Build Status](https://travis-ci.org/lytics/changepoint.svg?branch=master)](https://travis-ci.org/lytics/changepoint) [![GoDoc](https://godoc.org/github.com/lytics/changepoint?status.svg)](https://godoc.org/github.com/lytics/changepoint)

Lightweight bootstrap testing for changepoint analysis in [Go](https://golang.org).

## Purpose

Changepoint performs nonparametric causal inference for changes in time series data.  It separates a given time series into disjoint and adjacent sub-series to determine if there's that the structure of the time series process has changed.

Changepoint is designed to be free of any distributional assumptions, for use in any processes whose likelihood function is either unknown or dynamic.  It is location, scale and distribution free.

## Design

Changepoint creates a *reference* sub-series, which serves to provide information on the structural profile of the series.  It also creates a *candidate* sub-series which supplies evidence for or against the structural change.  

Changepoint assumes that both the *reference* and *candidate* subseries come from the same process under the same generative parameters.  Under this assumption, it performs Monte Carlo simulation (via bootstrap resampling) to simulate possible alternatives to the *candidate* subseries using data from the *reference* subseries.  The destination of each simulated walk is compared against the realized value from the *candidate* subseries, and the percentage of destinations considered as or more extreme (in terms of absolute deviation from the series) create the "p-value" for the test.

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

	"github.com/lytics/changepoint"
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
