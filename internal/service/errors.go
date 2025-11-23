package service

import "errors"

// Predefined errors for service layer following Go 1.13+ error handling best practices
var (
	// User errors
	ErrUserNotFound = errors.New("user not found")

	// Team errors
	ErrTeamNotFound      = errors.New("team not found")
	ErrTeamAlreadyExists = errors.New("team already exists")

	// PR errors
	ErrPRNotFound            = errors.New("PR not found")
	ErrPRAlreadyMerged       = errors.New("cannot reassign reviewer: PR is already merged")
	ErrReviewerNotAssigned   = errors.New("old reviewer is not assigned to this PR")
	ErrReviewerNotInTeam     = errors.New("reviewer is not in any team")
	ErrNoAvailableReviewers  = errors.New("no available reviewers in the team")
	ErrAuthorNotFound        = errors.New("author not found")
	ErrAuthorNotInTeam       = errors.New("author is not in any team")
	ErrInsufficientReviewers = errors.New("insufficient active reviewers in team")
	ErrCannotReviewOwnPR     = errors.New("author cannot review their own PR")
)
