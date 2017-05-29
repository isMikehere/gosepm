package test

import (
	"testing"
)

func Test_copySlice(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
		{name: "mike"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			copySlice()
		})
	}
}

func TestTestChan(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
		{name: "mike"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			TestChan()
		})
	}
}
