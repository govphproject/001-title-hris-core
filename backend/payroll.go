package main

// CalculateNet returns the net pay computed from gross, deductions and taxes.
func CalculateNet(gross, deductions, taxes float64) float64 {
	return gross - deductions - taxes
}
