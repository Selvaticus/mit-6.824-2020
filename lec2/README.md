# Lecture 2 - Threads and RPC

This lecture goes into why the need for threads to achieve concurrency and parallelism, it strengths and pitfalls.

Also describes why the choice of Golang was made.

[Video](https://pdos.csail.mit.edu/6.824/video/2.html) | [Notes](https://pdos.csail.mit.edu/6.824/notes/l-rpc.txt)

*Notes: Code provided as example by staff (not my code)*

## RPC

File `kv.go` shows an example given by the staff of a key/value store server and the use of RPC to store and retrieve values

```sh
# From the project root
$ go run lec2/kv.go

# Should see the output
Put(subject, 6.824) done
get(subject) -> 6.824 
```

## Threads

File `crawler.go` shows a couple of example of goroutines (Threads for Go), crawling a website (dummy structure) in parallel using goroutines and different sync methods, i.e channels and shared data + mutex.

```sh
# From the project root
$ go run lec2/crawler.go

# Output see the output
=== Serial===
found:   http://golang.org/
found:   http://golang.org/pkg/
missing: http://golang.org/cmd/
found:   http://golang.org/pkg/fmt/
found:   http://golang.org/pkg/os/
=== ConcurrentMutex ===
found:   http://golang.org/
missing: http://golang.org/cmd/
found:   http://golang.org/pkg/
found:   http://golang.org/pkg/os/
found:   http://golang.org/pkg/fmt/
=== ConcurrentChannel ===
found:   http://golang.org/
missing: http://golang.org/cmd/
found:   http://golang.org/pkg/
found:   http://golang.org/pkg/os/
found:   http://golang.org/pkg/fmt/
```

Note that there should be no duplicate URLs, i.e. URLs being parsed twice, in all implementations