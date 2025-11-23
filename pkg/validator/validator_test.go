package validator

import (
	"testing"

	"github.com/Rodjolo/pr-reviewer-service/pkg/dto"
)

func TestValidate_CreatePRRequest_Success(t *testing.T) {
	req := dto.CreatePRRequest{
		Title:    "Add new feature",
		AuthorID: 1,
	}

	err := Validate(&req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestValidate_CreatePRRequest_EmptyTitle(t *testing.T) {
	req := dto.CreatePRRequest{
		Title:    "",
		AuthorID: 1,
	}

	err := Validate(&req)
	if err == nil {
		t.Error("Expected validation error for empty title")
	}

	formatted := FormatValidationErrors(err)
	if formatted == "" {
		t.Error("Expected formatted error message")
	}
}

func TestValidate_CreatePRRequest_InvalidAuthorID(t *testing.T) {
	req := dto.CreatePRRequest{
		Title:    "Valid title",
		AuthorID: 0,
	}

	err := Validate(&req)
	if err == nil {
		t.Error("Expected validation error for invalid author_id")
	}

	formatted := FormatValidationErrors(err)
	if formatted == "" {
		t.Error("Expected formatted error message")
	}
}

func TestValidate_CreateUserRequest_Success(t *testing.T) {
	req := dto.CreateUserRequest{
		Name: "Alice",
	}

	err := Validate(&req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestValidate_CreateUserRequest_EmptyName(t *testing.T) {
	req := dto.CreateUserRequest{
		Name: "",
	}

	err := Validate(&req)
	if err == nil {
		t.Error("Expected validation error for empty name")
	}
}

func TestValidate_CreateTeamRequest_Success(t *testing.T) {
	req := dto.CreateTeamRequest{
		Name: "backend",
	}

	err := Validate(&req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestValidate_CreateTeamRequest_EmptyName(t *testing.T) {
	req := dto.CreateTeamRequest{
		Name: "",
	}

	err := Validate(&req)
	if err == nil {
		t.Error("Expected validation error for empty name")
	}
}

func TestValidate_AddMemberRequest_Success(t *testing.T) {
	req := dto.AddMemberRequest{
		UserID: 1,
	}

	err := Validate(&req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestValidate_AddMemberRequest_InvalidUserID(t *testing.T) {
	req := dto.AddMemberRequest{
		UserID: 0,
	}

	err := Validate(&req)
	if err == nil {
		t.Error("Expected validation error for invalid user_id")
	}
}

func TestValidate_ReassignRequest_Success(t *testing.T) {
	req := dto.ReassignRequest{
		OldReviewerID: 2,
	}

	err := Validate(&req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestValidate_ReassignRequest_InvalidReviewerID(t *testing.T) {
	req := dto.ReassignRequest{
		OldReviewerID: -1,
	}

	err := Validate(&req)
	if err == nil {
		t.Error("Expected validation error for invalid old_reviewer_id")
	}
}

func TestFormatValidationErrors(t *testing.T) {
	req := dto.CreatePRRequest{
		Title:    "",
		AuthorID: -1,
	}

	err := Validate(&req)
	if err == nil {
		t.Fatal("Expected validation error")
	}

	formatted := FormatValidationErrors(err)

	if formatted == "" {
		t.Error("Expected formatted error message, got empty string")
	}

	if len(formatted) < 10 {
		t.Errorf("Expected detailed error message, got: %s", formatted)
	}
}
