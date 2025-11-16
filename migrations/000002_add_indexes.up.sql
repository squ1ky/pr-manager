CREATE INDEX idx_users_team_name ON users (team_name);
CREATE INDEX idx_pull_requests_author_id ON pull_requests (author_id);
CREATE INDEX idx_pr_reviewers_reviewer_id_pull_request_id ON pull_request_reviewers (reviewer_id, pull_request_id);