package models

// Candidate represents an applicant in the recruitment system.
type Candidate struct {
    CandidateID string `bson:"candidate_id" json:"candidate_id"`
    Name        string `bson:"name" json:"name"`
    Email       string `bson:"email" json:"email"`
    Phone       string `bson:"phone,omitempty" json:"phone,omitempty"`
    ResumeURL   string `bson:"resume_url,omitempty" json:"resume_url,omitempty"`
    AppliedAt   int64  `bson:"applied_at,omitempty" json:"applied_at,omitempty"`
    Status      string `bson:"status,omitempty" json:"status,omitempty"`
    Version     int    `bson:"version,omitempty" json:"version,omitempty"`
}
