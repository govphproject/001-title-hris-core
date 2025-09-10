package models

// PayrollRecord represents a payroll entry for an employee.
type PayrollRecord struct {
    PayrollID  string  `bson:"payroll_id" json:"payroll_id"`
    EmployeeID string  `bson:"employee_id" json:"employee_id"`
    Period     string  `bson:"period" json:"period"` // e.g. 2025-08
    Gross      float64 `bson:"gross" json:"gross"`
    Deductions float64 `bson:"deductions" json:"deductions"`
    Taxes      float64 `bson:"taxes" json:"taxes"`
    Net        float64 `bson:"net" json:"net"`
    Currency   string  `bson:"currency,omitempty" json:"currency,omitempty"`
    CreatedAt  int64   `bson:"created_at,omitempty" json:"created_at,omitempty"`
    Version    int     `bson:"version,omitempty" json:"version,omitempty"`
}
