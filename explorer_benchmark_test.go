package main

import (
	"testing"
)

func BenchmarkGetFolderNotes(b *testing.B) {
	explorer := Explorer{
		BasePath: "testdata",
	}

	b.ResetTimer()
	for b.Loop() {
		_, err := explorer.getFolderNotes("")
		if err != nil {
			b.Fatalf("getFolderNotes() error = %v", err)
		}
	}
}

func BenchmarkGetFolderNotesParallel(b *testing.B) {
	explorer := Explorer{
		BasePath: "testdata",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := explorer.getFolderNotes("")
			if err != nil {
				b.Fatalf("getFolderNotes() error = %v", err)
			}
		}
	})
}
