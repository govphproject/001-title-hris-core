package models

// PerformanceReview represents a periodic evaluation.
type PerformanceReview struct {
    ReviewID    string `bson:"review_id" json:"review_id"`
    EmployeeID  string `bson:"employee_id" json:"employee_id"`
    Period      string `bson:"period" json:"period"`
    Score       int    `bson:"score" json:"score"`
    Comments    string `bson:"comments,omitempty" json:"comments,omitempty"`
    ReviewedBy  string `bson:"reviewed_by,omitempty" json:"reviewed_by,omitempty"`
    ReviewedAt  int64  `bson:"reviewed_at,omitempty" json:"reviewed_at,omitempty"`
    Version     int    `bson:"version,omitempty" json:"version,omitempty"`
}
