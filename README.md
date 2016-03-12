watrix
======

Caution
-------
Currently, the API is unstable because I'm experimenting various queries.

[![Build Status](https://travis-ci.org/sekineh/waveletTree.svg?branch=master)](https://travis-ci.org/sekineh/waveletTree)

watrix is a Go package for myriad array operations using wavelet matrix
(wavelet tree).

watrix stores a non-negative intger array V[0...n), 0 <= V[i] < s and
support almost all operations in O(log s) time (not depends on num) using
at most (n * log_2 s) bits plus small overheads for storing auxiually indices.


Usage
=====

[![GoDoc](https://godoc.org/github.com/sekineh/waveletTree?status.svg)](https://godoc.org/github.com/sekineh/waveletTree)

See godoc for reference.  It was originally folked from github.com/hillbig/waveletTree,
but compatibility is not maintained.

Benchmark
=========

New
---
- 2.8 GHz Xeon [in AWS]
- 32 cores (only a single thread is used)

Note: IgnoreLSBs version performs quite well.

	go test -v -bench . -benchmem

	{N = 10000000 is used in the tests below}   

	BenchmarkWM_Build-2    						1	11816507411 ns/op	10731534600 B/op	    7925 allocs/op

	BenchmarkWM_Lookup-2                	   50000	     24509 ns/op	       0 B/op	       0 allocs/op
	BenchmarkWM_Rank-2                  	   50000	     27384 ns/op	       0 B/op	       0 allocs/op
	BenchmarkWM_RangedRankIgnoreLSBs-2  	  100000	     22119 ns/op	       0 B/op	       0 allocs/op
	BenchmarkWM_RankLessThan-2          	   50000	     26901 ns/op	       0 B/op	       0 allocs/op

	BenchmarkWM_Select-2                	   30000	     43022 ns/op	       0 B/op	       0 allocs/op
	BenchmarkWM_RangedSelectIgnoreLSBs-2	   30000	     47325 ns/op	       0 B/op	       0 allocs/op

	BenchmarkWM_Quantile-2              	   50000	     23663 ns/op	       0 B/op	       0 allocs/op

	BenchmarkRaw_Lookup-2               	20000000	       124 ns/op	       0 B/op	       0 allocs/op

	BenchmarkRaw_Rank-2                 	     200	   5494451 ns/op	       0 B/op	       0 allocs/op
	BenchmarkRaw_Select-2               	     100	  17108476 ns/op	       0 B/op	       0 allocs/op
	BenchmarkRaw_Quantile-2             	      50	  20772400 ns/op	      32 B/op	       1 allocs/op


Old
---

- 1.7 GHz Intel Core i7
- OS X 10.9.2
- 8GB 1600 MHz DDR3
- go version go1.3 darwin/amd64

The results shows that RSDic operations require always
(almost) constant time with regard to the length and one's ratio.

	go test -bench=.

	// Build a watrix for an integer array of length 10^6 with s = 2^64
	BenchmarkWTBuild1M	       1	1455321650 ns/op
	// 1.455 micro sec per an interger

	// A watrix for an integer array of length 10M (10^7) with s = 2^64
	BenchmarkWTBuild10M	       1	1467061166 ns/op
	BenchmarkWTLookup10M	  100000	     29319 ns/op
	BenchmarkWTRank10M	  100000	     28278 ns/op
	BenchmarkWTSelect10M	   50000	     50250 ns/op
	BenchmarkWTQuantile10M	  100000	     28852 ns/op

	// An array []uint64 of length 10M (10^7) for comparison
	BenchmarkRawLookup10M	20000000	       109 ns/op
	BenchmarkRa
	wRank10M	     500	   4683822 ns/op
	BenchmarkRawSelect10M	     500	   6085992 ns/op
	BenchmarkRawQuantile10M	     100	  44362885 ns/op
