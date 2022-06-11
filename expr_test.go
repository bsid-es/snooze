package main

import (
	"testing"
	"time"
)

func BenchmarkParse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Parse("* */2 * * *")
	}
}

func BenchmarkNext(b *testing.B) {
	expr, err := Parse("0 * 2/2 * *")
	if err != nil {
		b.Fatal(err)
	}
	now := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		expr.Next(now)
	}
}
