package models

// BenefitPlan represents a set of benefits that can be assigned to employees.
type BenefitPlan struct {
    PlanID      string `bson:"plan_id" json:"plan_id"`
    Name        string `bson:"name" json:"name"`
    Description string `bson:"description,omitempty" json:"description,omitempty"`
    Version     int    `bson:"version,omitempty" json:"version,omitempty"`
}
