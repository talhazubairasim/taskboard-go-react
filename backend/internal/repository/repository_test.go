package repository

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"taskboard/internal/database"
	"taskboard/internal/models"
)

// Set TEST_DATABASE_URL to run these tests against a throwaway database.
func setup(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	pool, err := database.Connect(ctx, url, 3)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.MigrateUp(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `TRUNCATE tasks, users CASCADE`); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func str(s string) *string { return &s }

func TestUserRepository(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	r := NewUserRepository(pool)

	u, err := r.Create(ctx, &models.User{Email: " Alice@Example.com ", PasswordHash: "h", Name: "Alice"})
	if err != nil {
		t.Fatal(err)
	}
	if u.ID == "" || u.Email != "alice@example.com" || u.CreatedAt.IsZero() {
		t.Fatalf("created = %+v", u)
	}
	if _, err := r.Create(ctx, &models.User{Email: "alice@example.com", PasswordHash: "h", Name: "Dup"}); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("want duplicate, got %v", err)
	}

	got, err := r.GetByID(ctx, u.ID)
	if err != nil || got.Name != "Alice" {
		t.Fatalf("GetByID: %v %v", got, err)
	}
	if got, err := r.GetByEmail(ctx, "ALICE@example.com"); err != nil || got.ID != u.ID {
		t.Fatalf("GetByEmail: %v %v", got, err)
	}
	if _, err := r.GetByID(ctx, "00000000-0000-0000-0000-000000000000"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want not found, got %v", err)
	}
	if _, err := r.GetByID(ctx, "not-a-uuid"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("bad uuid: want not found, got %v", err)
	}

	exists, err := r.EmailExists(ctx, "alice@EXAMPLE.com")
	if err != nil || !exists {
		t.Fatal("EmailExists false for existing")
	}
	exists, _ = r.EmailExists(ctx, "bob@example.com")
	if exists {
		t.Fatal("EmailExists true for missing")
	}

	b, _ := r.Create(ctx, &models.User{Email: "bob@example.com", PasswordHash: "h", Name: "Bob"})
	list, err := r.List(ctx)
	if err != nil || len(list) != 2 {
		t.Fatalf("List: %d %v", len(list), err)
	}
	m, err := r.GetByIDs(ctx, []string{u.ID, b.ID})
	if err != nil || len(m) != 2 || m[b.ID].Name != "Bob" {
		t.Fatalf("GetByIDs: %v %v", m, err)
	}
	if m, err := r.GetByIDs(ctx, nil); err != nil || len(m) != 0 {
		t.Fatal("GetByIDs(nil) should be empty")
	}
	if _, err := r.GetByIDs(ctx, []string{"bad"}); err == nil {
		t.Fatal("GetByIDs with bad uuid should error")
	}

	upd, err := r.Update(ctx, u.ID, models.UserUpdate{Name: str("Alicia"), Avatar: str("http://a/x.png")})
	if err != nil || upd.Name != "Alicia" || *upd.Avatar != "http://a/x.png" {
		t.Fatalf("Update: %+v %v", upd, err)
	}
	upd, _ = r.Update(ctx, u.ID, models.UserUpdate{})
	if upd.Name != "Alicia" {
		t.Fatal("empty update changed name")
	}
	if _, err := r.Update(ctx, "00000000-0000-0000-0000-000000000000", models.UserUpdate{Name: str("x")}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("update missing: %v", err)
	}
}

func TestTaskRepository(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	users := NewUserRepository(pool)
	r := NewTaskRepository(pool)

	alice, _ := users.Create(ctx, &models.User{Email: "a@x.com", PasswordHash: "h", Name: "A"})
	bob, _ := users.Create(ctx, &models.User{Email: "b@x.com", PasswordHash: "h", Name: "B"})
	due := time.Now().Add(48 * time.Hour).UTC().Truncate(time.Second)

	t1, err := r.Create(ctx, &models.Task{Title: "One", Description: str("d"), Status: models.StatusTodo,
		Priority: models.PriorityHigh, CreatedByID: alice.ID, AssignedToID: &bob.ID, DueDate: &due})
	if err != nil {
		t.Fatal(err)
	}
	t2, _ := r.Create(ctx, &models.Task{Title: "Two", Status: models.StatusDone, Priority: models.PriorityLow, CreatedByID: bob.ID})

	if _, err := r.Create(ctx, &models.Task{Title: "Bad", Status: models.StatusTodo, Priority: models.PriorityLow,
		CreatedByID: alice.ID, AssignedToID: str("00000000-0000-0000-0000-000000000000")}); !errors.Is(err, ErrInvalidReference) {
		t.Fatalf("want invalid reference, got %v", err)
	}

	got, err := r.GetByID(ctx, t1.ID)
	if err != nil || got.Title != "One" || *got.AssignedToID != bob.ID || !got.DueDate.Equal(due) {
		t.Fatalf("GetByID: %+v %v", got, err)
	}
	if _, err := r.GetByID(ctx, "00000000-0000-0000-0000-000000000000"); !errors.Is(err, ErrNotFound) {
		t.Fatal("missing task should be ErrNotFound")
	}

	all, _ := r.List(ctx, models.TaskFilter{})
	if len(all) != 2 {
		t.Fatalf("List all = %d", len(all))
	}
	for _, tc := range []struct {
		f    models.TaskFilter
		want int
	}{
		{models.TaskFilter{Status: models.StatusTodo}, 1},
		{models.TaskFilter{Priority: models.PriorityLow}, 1},
		{models.TaskFilter{AssignedToID: bob.ID}, 1},
		{models.TaskFilter{CreatedByID: bob.ID}, 1},
		{models.TaskFilter{Status: models.StatusDone, CreatedByID: alice.ID}, 0},
	} {
		got, err := r.List(ctx, tc.f)
		if err != nil || len(got) != tc.want {
			t.Errorf("filter %+v: got %d err %v", tc.f, len(got), err)
		}
	}
	if _, err := r.List(ctx, models.TaskFilter{AssignedToID: "bad-uuid"}); err == nil {
		t.Fatal("bad uuid filter should error")
	}

	newDue := due.Add(time.Hour)
	upd, err := r.Update(ctx, t1.ID, models.TaskUpdate{Title: str("One!"), Description: str("dd"),
		Status: str(models.StatusInProgress), Priority: str(models.PriorityUrgent), AssignedToID: &alice.ID, DueDate: &newDue})
	if err != nil || upd.Title != "One!" || *upd.Description != "dd" || upd.Status != models.StatusInProgress ||
		upd.Priority != models.PriorityUrgent || *upd.AssignedToID != alice.ID || !upd.DueDate.Equal(newDue) {
		t.Fatalf("Update: %+v %v", upd, err)
	}
	if !upd.UpdatedAt.After(upd.CreatedAt) && !upd.UpdatedAt.Equal(upd.CreatedAt) {
		t.Fatal("updated_at not maintained")
	}
	upd, err = r.Update(ctx, t1.ID, models.TaskUpdate{ClearAssignee: true, ClearDueDate: true, ClearDescription: true})
	if err != nil || upd.AssignedToID != nil || upd.DueDate != nil || upd.Description != nil {
		t.Fatalf("clear: %+v %v", upd, err)
	}
	if upd, err := r.Update(ctx, t1.ID, models.TaskUpdate{}); err != nil || upd.Title != "One!" {
		t.Fatal("empty update should return the task unchanged")
	}
	if _, err := r.Update(ctx, t1.ID, models.TaskUpdate{Status: str("NOPE")}); err == nil {
		t.Fatal("check constraint should reject bad status")
	}
	if _, err := r.Update(ctx, "00000000-0000-0000-0000-000000000000", models.TaskUpdate{Title: str("x")}); !errors.Is(err, ErrNotFound) {
		t.Fatal("update missing should be ErrNotFound")
	}

	if err := r.Delete(ctx, t2.ID); err != nil {
		t.Fatal(err)
	}
	if err := r.Delete(ctx, t2.ID); !errors.Is(err, ErrNotFound) {
		t.Fatal("double delete should be ErrNotFound")
	}
	if err := r.Delete(ctx, "bad"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("delete bad id: %v", err)
	}
}

func TestClosedPoolErrors(t *testing.T) {
	pool := setup(t)
	ctx := context.Background()
	users := NewUserRepository(pool)
	tasks := NewTaskRepository(pool)
	pool.Close()
	if _, err := users.List(ctx); err == nil {
		t.Error("users.List")
	}
	if _, err := users.EmailExists(ctx, "x"); err == nil {
		t.Error("EmailExists")
	}
	if _, err := users.Create(ctx, &models.User{Email: "x"}); err == nil {
		t.Error("users.Create")
	}
	if _, err := users.GetByIDs(ctx, []string{"x"}); err == nil {
		t.Error("GetByIDs")
	}
	if _, err := tasks.List(ctx, models.TaskFilter{}); err == nil {
		t.Error("tasks.List")
	}
	if _, err := tasks.Create(ctx, &models.Task{}); err == nil {
		t.Error("tasks.Create")
	}
	if err := tasks.Delete(ctx, "x"); err == nil {
		t.Error("tasks.Delete")
	}
}

func TestTranslatePassesThroughOtherErrors(t *testing.T) {
	e := errors.New("boom")
	if translate(e) != e {
		t.Fatal("non-pg errors must pass through")
	}
}
