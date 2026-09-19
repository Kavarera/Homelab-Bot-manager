package domain

import (
	"testing"
	"time"
)

func TestGenerateInvoiceNumber(t *testing.T) {
	issueDate := time.Date(2026, time.September, 19, 10, 0, 0, 0, time.UTC)
	invNum := GenerateInvoiceNumber(1, issueDate)

	expected := "INV/1/2026/09/19"
	if invNum != expected {
		t.Errorf("expected invoice number %q, got %q", expected, invNum)
	}

	invNum2 := GenerateInvoiceNumber(105, time.Date(2027, time.January, 5, 0, 0, 0, 0, time.UTC))
	expected2 := "INV/105/2027/01/05"
	if invNum2 != expected2 {
		t.Errorf("expected invoice number %q, got %q", expected2, invNum2)
	}
}

func TestCalculateDueDate_Weekday(t *testing.T) {
	// Monday 2026-09-07 + 14 days = Monday 2026-09-21
	issueDate := time.Date(2026, time.September, 7, 0, 0, 0, 0, time.UTC)
	dueDate := CalculateDueDate(issueDate)

	expected := time.Date(2026, time.September, 21, 0, 0, 0, 0, time.UTC)
	if !dueDate.Equal(expected) {
		t.Errorf("expected due date %v, got %v (Weekday: %v)", expected, dueDate, dueDate.Weekday())
	}
}

func TestCalculateDueDate_FallsOnSaturday(t *testing.T) {
	// Saturday 2026-09-05 + 14 days = Saturday 2026-09-19 -> Rolls to Monday 2026-09-21
	issueDate := time.Date(2026, time.September, 5, 0, 0, 0, 0, time.UTC) // Saturday
	dueDate := CalculateDueDate(issueDate)

	expected := time.Date(2026, time.September, 21, 0, 0, 0, 0, time.UTC) // Monday
	if !dueDate.Equal(expected) {
		t.Errorf("expected due date rolled to Monday %v, got %v (Weekday: %v)", expected, dueDate, dueDate.Weekday())
	}
	if dueDate.Weekday() != time.Monday {
		t.Errorf("expected weekday Monday, got %v", dueDate.Weekday())
	}
}

func TestCalculateDueDate_FallsOnSunday(t *testing.T) {
	// Sunday 2026-09-06 + 14 days = Sunday 2026-09-20 -> Rolls to Monday 2026-09-21
	issueDate := time.Date(2026, time.September, 6, 0, 0, 0, 0, time.UTC) // Sunday
	dueDate := CalculateDueDate(issueDate)

	expected := time.Date(2026, time.September, 21, 0, 0, 0, 0, time.UTC) // Monday
	if !dueDate.Equal(expected) {
		t.Errorf("expected due date rolled to Monday %v, got %v (Weekday: %v)", expected, dueDate, dueDate.Weekday())
	}
	if dueDate.Weekday() != time.Monday {
		t.Errorf("expected weekday Monday, got %v", dueDate.Weekday())
	}
}
