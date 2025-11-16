package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/F3dosik/PRS.git/internal/models/api"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(dsn string) (*Storage, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	storage := &Storage{
		db: db,
	}

	return storage, nil
}

func (s *Storage) UpdateTeam(ctx context.Context, team *api.Team) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var teamID uuid.UUID

	err = tx.QueryRowContext(ctx, `
		INSERT INTO teams (name)
		VALUES ($1)
		RETURNING id
		`, team.TeamName).Scan(&teamID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return api.NewAPIError(api.ErrTeamExist, "team_name already exists")
		}

		return fmt.Errorf("insert team: %w", err)
	}

	for _, member := range team.Members {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO users (id, name, is_active, team_id)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO UPDATE
			SET name = EXCLUDED.name, 
				is_active = EXCLUDED.is_active,
				team_id = EXCLUDED.team_id
		`, member.UserID, member.Username, member.IsActive, teamID)
		if err != nil {
			return fmt.Errorf("upsert user: %w", err)
		}
	}

	return tx.Commit()
}

func (s *Storage) GetTeam(ctx context.Context, teamName string) (*api.Team, error) {
	var teamID uuid.UUID
	var team api.Team

	err := s.db.QueryRowContext(ctx, `
		SELECT id, name FROM teams
		WHERE name = $1
	`, teamName).Scan(&teamID, &team.TeamName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, api.NewAPIError(api.ErrNotFound, "team not found")
		}
		return nil, fmt.Errorf("query team: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, is_active FROM users
		WHERE team_id = $1
	`, teamID)
	if err != nil {
		return nil, fmt.Errorf("query team members: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var member api.TeamMember
		if err = rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			return nil, fmt.Errorf("scan member: %w", err)
		}
		team.Members = append(team.Members, member)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return &team, nil
}

func (s *Storage) SetIsActive(ctx context.Context, userID uuid.UUID, isActive bool) (*api.User, error) {
	var username string
	var teamname *string
	err := s.db.QueryRowContext(ctx, `
		SELECT u.name, t.name
		FROM users u
		LEFT JOIN teams t ON u.team_id = t.id
		WHERE u.id = $1
	`, userID).Scan(&username, &teamname)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, api.NewAPIError(api.ErrNotFound, "user not found")
		}
		return nil, fmt.Errorf("query user: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE users
		SET is_active = $1
		WHERE id = $2
	`, isActive, userID)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	user := api.User{
		UserID:   userID,
		Username: username,
		TeamName: teamname,
		IsActive: isActive,
	}

	return &user, nil
}

func (s *Storage) PullRequestCreate(ctx context.Context, prID, authorID uuid.UUID, prName string) (*api.PullRequest, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var teamID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		SELECT team_id FROM users
		WHERE id = $1
	`, authorID).Scan(&teamID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, api.NewAPIError(api.ErrNotFound, "author not found")
		}
		return nil, fmt.Errorf("query author team: %w", err)
	}

	if teamID == uuid.Nil {
		return nil, api.NewAPIError(api.ErrNotFound, "author has no team")
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO pull_request (id, title, author_id)
		VALUES ($1, $2, $3)
	`, prID, prName, authorID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, api.NewAPIError(api.ErrPRExist, "PR id already exist")
		}
		return nil, fmt.Errorf("insert pr: %w", err)
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT id FROM users
		WHERE team_id = $1
			AND is_active = true
			AND id <> $2
		ORDER BY RANDOM()
		LIMIT 2
	`, teamID, authorID)
	if err != nil {
		return nil, fmt.Errorf("query reviewers: %w", err)
	}
	defer rows.Close()

	var reviewers []uuid.UUID
	for rows.Next() {
		var reviewer uuid.UUID
		if err = rows.Scan(&reviewer); err != nil {
			return nil, fmt.Errorf("scan reviwer: %w", err)
		}
		reviewers = append(reviewers, reviewer)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	var reviewer1ID, reviewer2ID *uuid.UUID

	if len(reviewers) > 0 {
		reviewer1ID = &reviewers[0]
	}
	if len(reviewers) > 1 {
		reviewer2ID = &reviewers[1]
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE pull_request
		SET reviewer1_id = $1,
			reviewer2_id = $2,
			need_more_reviewers = $3
		WHERE id = $4
	`, reviewer1ID, reviewer2ID, len(reviewers) < 2, prID)
	if err != nil {
		return nil, fmt.Errorf("update pull request: %w", err)
	}

	pr := api.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            api.StatusOpen,
		AssignedReviewers: reviewers,
	}

	return &pr, tx.Commit()
}

func (s *Storage) PullRequestMerge(ctx context.Context, prID uuid.UUID) (*api.PullRequest, error) {
	var (
		title       string
		authorID    uuid.UUID
		status      api.PRStatus
		reviewer1ID *uuid.UUID
		reviewer2ID *uuid.UUID
		createdAt   time.Time
		mergedAt    *time.Time
	)

	err := s.db.QueryRowContext(ctx, `
		SELECT title, author_id, status, reviewer1_id, reviewer2_id, created_at, merged_at
		FROM pull_request
		WHERE id = $1
	`, prID).Scan(
		&title, &authorID, &status,
		&reviewer1ID, &reviewer2ID,
		&createdAt, &mergedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, api.NewAPIError(api.ErrNotFound, "pull request not found")
		}
		return nil, fmt.Errorf("query pull request: %w", err)
	}

	pr := &api.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   title,
		AuthorID:          authorID,
		Status:            api.StatusMerged,
		AssignedReviewers: makeReviewers(reviewer1ID, reviewer2ID),
		CreatedAt:         createdAt,
		MergedAt:          mergedAt,
	}

	if status == api.StatusMerged {
		return pr, nil
	}

	err = s.db.QueryRowContext(ctx, `
		UPDATE pull_request
		SET status = 'MERGED',
			merged_at = now()
		WHERE id = $1
		RETURNING merged_at
	`, prID).Scan(&mergedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, api.NewAPIError(api.ErrNotFound, "pull request not found")
		}
		return nil, fmt.Errorf("update pull_request: %w", err)
	}

	pr.Status = api.StatusMerged
	pr.MergedAt = mergedAt

	return pr, nil
}

func (s *Storage) PullRequestReassign(ctx context.Context, prID, oldUserID uuid.UUID) (*api.PullRequestReassignResponse, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var exist bool
	err = tx.QueryRowContext(ctx, `
		SELECT 
		EXISTS (SELECT 1 FROM pull_request WHERE id = $1)
		AND
		EXISTS (SELECT 1 FROM users WHERE id = $2)
	`, prID, oldUserID).Scan(&exist)
	if err != nil {
		return nil, fmt.Errorf("check existence: %w", err)
	}
	if !exist {
		return nil, api.NewAPIError(api.ErrNotFound, "pull_request_id or old_user_id not found")
	}

	var status api.PRStatus
	err = tx.QueryRowContext(ctx, `
		SELECT status FROM pull_request
		WHERE id = $1
	`, prID).Scan(&status)
	if err != nil {
		return nil, fmt.Errorf("query pr status: %w", err)
	}
	if status == api.StatusMerged {
		return nil, api.NewAPIError(api.ErrPRMerged, "cannot reassign on merged PR")
	}

	var teamID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		SELECT team_id FROM users WHERE id = (
			SELECT author_id FROM pull_request WHERE id = $1
		)
	`, prID).Scan(&teamID)
	if err != nil {
		return nil, fmt.Errorf("query team_id: %w", err)
	}

	var newUserID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		SELECT id FROM users
		WHERE is_active = true
			AND team_id = $1
			AND id <> $2
		ORDER BY RANDOM()
		LIMIT 1
	`, teamID, oldUserID).Scan(&newUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, api.NewAPIError(api.ErrNoCandidate, "no active replacement candidate in team")
		}
		return nil, fmt.Errorf("query users: %w", err)
	}

	var (
		title       string
		authorID    uuid.UUID
		reviewer1ID *uuid.UUID
		reviewer2ID *uuid.UUID
	)

	err = tx.QueryRowContext(ctx, `
		UPDATE pull_request
		SET
			reviewer1_id = CASE WHEN reviewer1_id = $1 THEN $2 ELSE reviewer1_id END,
			reviewer2_id = CASE WHEN reviewer2_id = $1 THEN $2 ELSE reviewer2_id END
		WHERE id = $3
		RETURNING title, author_id, reviewer1_id, reviewer2_id
	`, oldUserID, newUserID, prID).Scan(&title, &authorID, &reviewer1ID, &reviewer2ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, api.NewAPIError(api.ErrNotAssigned, "reviewer is not assigned to this PR")
		}
		return nil, fmt.Errorf("update pull request: %w", err)
	}

	pr := api.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   title,
		AuthorID:          authorID,
		Status:            status,
		AssignedReviewers: makeReviewers(reviewer1ID, reviewer2ID),
	}
	prResponse := &api.PullRequestReassignResponse{
		PullRequest: pr,
		ReplacedBy:  newUserID,
	}
	return prResponse, tx.Commit()
}

func (s *Storage) GetReview(ctx context.Context, userID uuid.UUID) (*api.GetReviewResponse, error) {
	var exist bool
	err := s.db.QueryRowContext(ctx, `
		SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)
	`, userID).Scan(&exist)
	if err != nil {
		return nil, fmt.Errorf("query exist user_id: %w", err)
	}
	if !exist {
		return nil, api.NewAPIError(api.ErrNotFound, "user not found")
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, title, author_id, status FROM pull_request
		WHERE reviewer1_id = $1 OR reviewer2_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("query pull_request: %w", err)
	}
	defer rows.Close()

	var prs []api.PullRequestShort
	for rows.Next() {
		var prShort api.PullRequestShort
		err = rows.Scan(&prShort.PullRequestID, &prShort.PullRequestName, &prShort.AuthorID, &prShort.Status)
		if err != nil {
			return nil, fmt.Errorf("scan pull_request_short: %w", err)
		}
		prs = append(prs, prShort)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	grr := &api.GetReviewResponse{
		UserID:       userID,
		PullRequests: prs,
	}

	return grr, nil
}

func (s *Storage) GetStats(ctx context.Context) (*api.StatsResponse, error) {
	stats := &api.StatsResponse{
		ReviewAssignments: make(map[string]int),
	}

	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE status = 'OPEN')
		FROM pull_request
	`).Scan(&stats.TotalPR, &stats.OpenPR)
	if err != nil {
		return nil, fmt.Errorf("query total PR: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT u.name, COUNT(*) AS total_reviews
		FROM (
			SELECT reviewer1_id AS reviewer_id FROM pull_request WHERE reviewer1_id IS NOT NULL
			UNION ALL
			SELECT reviewer2_id AS reviewer_id FROM pull_request WHERE reviewer2_id IS NOT NULL
		) sub
		JOIN users u ON u.id = sub.reviewer_id
		GROUP BY u.name
		ORDER BY total_reviews DESC;
	`)
	if err != nil {
		return nil, fmt.Errorf("query review assignments: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var userName string
		var count int
		if err := rows.Scan(&userName, &count); err != nil {
			return nil, fmt.Errorf("scan review assignment: %w", err)
		}
		stats.ReviewAssignments[userName] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return stats, nil
}

func makeReviewers(reviewer1ID, reviewer2ID *uuid.UUID) []uuid.UUID {
	reviewers := make([]uuid.UUID, 0, 2)
	if reviewer1ID != nil {
		reviewers = append(reviewers, *reviewer1ID)
	}
	if reviewer2ID != nil {
		reviewers = append(reviewers, *reviewer2ID)
	}
	return reviewers
}
