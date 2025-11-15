CREATE TABLE teams
(
    name       VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE users
(
    id         VARCHAR(255) PRIMARY KEY,
    username   VARCHAR(255) NOT NULL,
    team_name  VARCHAR(255) REFERENCES teams (name) ON DELETE SET NULL,
    is_active  BOOLEAN   DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE pull_requests
(
    id         VARCHAR(255) PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    author_id  VARCHAR(255) NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status     VARCHAR(20)  NOT NULL CHECK (status in ('OPEN', 'MERGED')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    merged_at  TIMESTAMP    NULL
);

CREATE TABLE pull_request_reviewers
(
    pull_request_id VARCHAR(255) NOT NULL REFERENCES pull_requests (id) ON DELETE CASCADE,
    reviewer_id     VARCHAR(255) NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    PRIMARY KEY (pull_request_id, reviewer_id)
);