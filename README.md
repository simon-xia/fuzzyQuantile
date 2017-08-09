# FuzzyQuantile

[![GoDoc](https://godoc.org/github.com/simon-xia/fuzzyQuantile?status.png)](https://godoc.org/github.com/simon-xia/fuzzyQuantile)
[![Go Report Card](https://goreportcard.com/badge/github.com/simon-xia/fuzzyQuantile)](https://goreportcard.com/report/github.com/simon-xia/fuzzyQuantile)

High performance quantile estimation(e.g. 90th, 95th, 99th) over streaming data, with user defined reasonable error (e.g 0.1%). 

This is an implementation of the algorithm presented in [Cormode, Korn, Muthukrishnan, and Srivastava. "Effective Computation of Biased Quantiles over Data Streams"](https://www.cs.rutgers.edu/~muthu/bquant.pdf) in ICDE 2005.


# Install

```
go get github.com/simon-xia/fuzzyQuantile
```


# Usage


This example show target quantile estimation. Given a set of Quantiles, each Quantile instance repsent a pair (quantile, error) which means expected quantile value with the error. And query will give the result quantile value corresponding error.

```go
	testQuantiles := []Quantile{
		NewQuantile(0.5, 0.01),
		NewQuantile(0.8, 0.001),
		NewQuantile(0.95, 0.0001),
	}

	fq := NewFuzzyQuantile(&FuzzyQuantileConf{Quantiles: testQuantiles})

	// valueChan repsent a data stream source
	valueChan := make(chan float64)
	for v := range valueChan {
		fq.Insert(v)
    }
    
    // valueChan close at other place

	v, er := fq.Query(0.8)
	if er != nil {
		// handle error
	}
	log.Printf("success 80th percentile value %v", v)
```

For other usage, check the document or testcase in source code

# TODO

- [ ] RBTree Storage Impl
- [ ] More graceful log
