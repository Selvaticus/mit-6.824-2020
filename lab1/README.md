# Lab 1: MapReduce

## Introduction

We will build a MapReduce system. We will implement a worker process that calls application Map and Reduce functions and handles reading and writing files, and a master process that hands out tasks to workers and copes with failed workers.

We will be building something similar to the [MapReduce paper](http://research.google.com/archive/mapreduce-osdi04.pdf).

### Notes

This document only has summary of directions given for this lab, for full information about this lab, by the teaching staff, go [here](https://pdos.csail.mit.edu/6.824/labs/lab-mr.html)

## Getting Started

A simple sequential mapreduce implementation is supplied in `main/mrsequential.go`. It runs the maps and reduces one at a time, in a single process. Also provided are a couple of MapReduce applications: word-count in `mrapps/wc.go`, and a text indexer in `mrapps/indexer.go`. You can run word count sequentially as follows:

```sh
$ cd lab1/main
$ go build -buildmode=plugin mrapps/wc.go
$ rm mr-out*
$ go run mrsequential.go wc.so pg*.txt
$ more mr-out-0
A 509
ABOUT 2
ACT 8
...
```

`mrsequential.go` leaves its output in the file `mr-out-0`. The input is from the text files named `pg-xxx.txt`.

From the staff:
> Feel free to borrow code from `mrsequential.go`. You should also have a look at `mrapps/wc.go` to see what MapReduce application code looks like

