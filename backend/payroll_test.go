package main

import "testing"

func TestCalculateNet(t *testing.T) {
    gross := 2000.0
    deductions := 150.0
    taxes := 250.0
    got := CalculateNet(gross, deductions, taxes)
    want := 1600.0
    if got != want {
        t.Fatalf("CalculateNet = %v; want %v", got, want)
    }
}
