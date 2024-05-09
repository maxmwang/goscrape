package main

import "testing"

// go test -bench . --benchmem
func BenchmarkMain(b *testing.B) {
	main()
}
