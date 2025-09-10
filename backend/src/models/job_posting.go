package models

// JobPosting represents an open position.
type JobPosting struct {
    JobID       string `bson:"job_id" json:"job_id"`
    Title       string `bson:"title" json:"title"`
    Department  string `bson:"department,omitempty" json:"department,omitempty"`
    Description string `bson:"description,omitempty" json:"description,omitempty"`
    Location    string `bson:"location,omitempty" json:"location,omitempty"`
    PostedAt    int64  `bson:"posted_at,omitempty" json:"posted_at,omitempty"`
    ClosedAt    *int64 `bson:"closed_at,omitempty" json:"closed_at,omitempty"`
    Version     int    `bson:"version,omitempty" json:"version,omitempty"`
}
