package models

// TrainingRecord represents a training event and an employee's registration.
type TrainingRecord struct {
    TrainingID   string `bson:"training_id" json:"training_id"`
    Title        string `bson:"title" json:"title"`
    Description  string `bson:"description,omitempty" json:"description,omitempty"`
    Date         string `bson:"date,omitempty" json:"date,omitempty"`
    Location     string `bson:"location,omitempty" json:"location,omitempty"`
    RegisteredBy []string `bson:"registered_by,omitempty" json:"registered_by,omitempty"`
    Version      int    `bson:"version,omitempty" json:"version,omitempty"`
}
