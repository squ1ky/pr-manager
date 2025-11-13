package entity

import "time"

type PullRequestStatus string

const (
	PullRequestStatusOpen   PullRequestStatus = "OPEN"
	PullRequestStatusMerged PullRequestStatus = "MERGED"
)

type PullRequest struct {
	ID        string            `json:"pull_request_id" db:"id"`
	Name      string            `json:"pull_request_name" db:"name"`
	AuthorID  string            `json:"author_id" db:"author_id"`
	Status    PullRequestStatus `json:"status" db:"status"`
	CreatedAt time.Time         `json:"created_at" db:"created_at"`
	MergedAt  *time.Time        `json:"merged_at" db:"merged_at"`
}

type PullRequestReviewer struct {
	PullRequestID string `db:"pull_request_id"`
	ReviewerID    string `db:"reviewer_id"`
}
