package models

import "testing"

func TestValidStatusAndPriority(t *testing.T) {
	for _, s := range []string{StatusTodo, StatusInProgress, StatusReview, StatusDone} {
		if !ValidStatus(s) {
			t.Errorf("%s should be valid", s)
		}
	}
	if ValidStatus("todo") || ValidStatus("") {
		t.Error("lowercase/empty status must be invalid")
	}
	for _, p := range []string{PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent} {
		if !ValidPriority(p) {
			t.Errorf("%s should be valid", p)
		}
	}
	if ValidPriority("CRITICAL") {
		t.Error("unknown priority must be invalid")
	}
}

func TestTaskUpdateEmpty(t *testing.T) {
	if !(TaskUpdate{}).Empty() {
		t.Fatal("zero update should be empty")
	}
	s := "DONE"
	if (TaskUpdate{Status: &s}).Empty() {
		t.Fatal("update with status is not empty")
	}
	if (TaskUpdate{ClearAssignee: true}).Empty() {
		t.Fatal("clear assignee is not empty")
	}
}
