package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"taskboard/internal/auth"
	"taskboard/internal/memstore"
	"taskboard/internal/models"
	"taskboard/internal/pubsub"
	"taskboard/internal/repository"
)

func init() { auth.BcryptCost = bcrypt.MinCost }

const goodPW = "Str0ng!Pass"

func sp(s string) *string { return &s }

type env struct {
	users *memstore.Users
	tasks *memstore.Tasks
	cache *memstore.Cache
	pub   *memstore.Recorder
	jwt   *auth.JWTManager
	auth  *AuthService
	user  *UserService
	task  *TaskService
}

func newEnv() *env {
	e := &env{users: memstore.NewUsers(), tasks: memstore.NewTasks(), cache: memstore.NewCache(), pub: &memstore.Recorder{},
		jwt: auth.NewJWTManager("s", "r")}
	e.auth = NewAuthService(e.users, e.jwt)
	e.user = NewUserService(e.users, e.cache, time.Minute)
	e.task = NewTaskService(e.tasks, e.users, e.cache, e.pub, time.Minute)
	return e
}

func (e *env) register(t *testing.T, email string) *models.User {
	t.Helper()
	r, err := e.auth.Register(context.Background(), email, goodPW, "User "+email)
	if err != nil {
		t.Fatal(err)
	}
	return r.User
}

func isValidation(err error) bool {
	var v *ValidationError
	return errors.As(err, &v)
}

// ---------- auth ----------

func TestRegisterLoginRefresh(t *testing.T) {
	e := newEnv()
	ctx := context.Background()
	r, err := e.auth.Register(ctx, " Alice@Example.COM ", goodPW, " Alice ")
	if err != nil {
		t.Fatal(err)
	}
	if r.User.Email != "alice@example.com" || r.User.Name != "Alice" || r.Token == "" || r.RefreshToken == "" {
		t.Fatalf("register result %+v", r)
	}
	if _, err := e.auth.Register(ctx, "alice@example.com", goodPW, "A"); !errors.Is(err, ErrConflict) {
		t.Fatalf("duplicate: %v", err)
	}

	l, err := e.auth.Login(ctx, "alice@example.com", goodPW)
	if err != nil || l.User.ID != r.User.ID {
		t.Fatalf("login: %v", err)
	}
	if _, err := e.auth.Login(ctx, "alice@example.com", "Wrong!Pass1"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("wrong pw: %v", err)
	}
	if _, err := e.auth.Login(ctx, "nobody@example.com", goodPW); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("unknown user: %v", err)
	}

	rr, err := e.auth.Refresh(ctx, r.RefreshToken)
	if err != nil || rr.User.ID != r.User.ID || rr.RefreshToken == r.RefreshToken {
		t.Fatalf("refresh: %v", err)
	}
	if _, err := e.auth.Refresh(ctx, r.Token); !errors.Is(err, ErrUnauthenticated) {
		t.Fatal("access token accepted as refresh token")
	}
	// refresh for a deleted user
	ghost, _ := e.jwt.GenerateRefreshToken("missing", "x")
	if _, err := e.auth.Refresh(ctx, ghost); !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("ghost refresh: %v", err)
	}
}

func TestRegisterValidation(t *testing.T) {
	e := newEnv()
	ctx := context.Background()
	for _, c := range []struct{ email, pw, name string }{
		{"not-an-email", goodPW, "A"},
		{"Bob <bob@x.com>", goodPW, "A"},
		{"a@b.com", goodPW, "   "},
		{"a@b.com", "weak", "A"},
	} {
		if _, err := e.auth.Register(ctx, c.email, c.pw, c.name); !isValidation(err) {
			t.Errorf("%+v: want validation error, got %v", c, err)
		}
	}
}

func TestAuthStoreFailures(t *testing.T) {
	ctx := context.Background()

	e := newEnv()
	e.users.Fail["EmailExists"] = true
	if _, err := e.auth.Register(ctx, "a@b.com", goodPW, "A"); !errors.Is(err, memstore.ErrBoom) {
		t.Fatal("EmailExists failure not returned")
	}

	e = newEnv()
	e.users.Fail["Create"] = true
	if _, err := e.auth.Register(ctx, "a@b.com", goodPW, "A"); !errors.Is(err, memstore.ErrBoom) {
		t.Fatal("Create failure not returned")
	}

	e = newEnv()
	e.users.Fail["GetByEmail"] = true
	if _, err := e.auth.Login(ctx, "a@b.com", goodPW); !errors.Is(err, memstore.ErrBoom) {
		t.Fatal("GetByEmail failure not returned")
	}

	e = newEnv()
	r, _ := e.auth.Register(ctx, "a@b.com", goodPW, "A")
	e.users.Fail["GetByID"] = true
	if _, err := e.auth.Refresh(ctx, r.RefreshToken); !errors.Is(err, memstore.ErrBoom) {
		t.Fatal("GetByID failure not returned")
	}

	auth.BcryptCost = 99
	defer func() { auth.BcryptCost = bcrypt.MinCost }()
	if _, err := newEnv().auth.Register(ctx, "c@d.com", goodPW, "C"); err == nil {
		t.Fatal("hash failure not returned")
	}
}

// ---------- users ----------

func TestUserService(t *testing.T) {
	e := newEnv()
	ctx := context.Background()
	u := e.register(t, "u@x.com")

	got, err := e.user.Get(ctx, u.ID)
	if err != nil || got.Email != "u@x.com" {
		t.Fatal(err)
	}
	hits := e.cache.Hits
	if _, err := e.user.Get(ctx, u.ID); err != nil || e.cache.Hits != hits+1 {
		t.Fatal("second Get should be a cache hit")
	}
	if _, err := e.user.Get(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatal("missing user")
	}

	list, err := e.user.List(ctx)
	if err != nil || len(list) != 1 {
		t.Fatal("list")
	}

	upd, err := e.user.UpdateProfile(ctx, u.ID, models.UserUpdate{Name: sp("  New "), Avatar: sp("https://img/x.png")})
	if err != nil || upd.Name != "New" || *upd.Avatar != "https://img/x.png" {
		t.Fatalf("update: %+v %v", upd, err)
	}
	// cache was invalidated
	got, _ = e.user.Get(ctx, u.ID)
	if got.Name != "New" {
		t.Fatal("stale cache after profile update")
	}
	if upd, err := e.user.UpdateProfile(ctx, u.ID, models.UserUpdate{Avatar: sp("")}); err != nil || *upd.Avatar != "" {
		t.Fatal("clearing avatar failed")
	}
	for _, bad := range []models.UserUpdate{
		{Name: sp("")},
		{Avatar: sp("javascript:alert(1)")},
		{Avatar: sp("https://")},
	} {
		if _, err := e.user.UpdateProfile(ctx, u.ID, bad); !isValidation(err) {
			t.Errorf("%+v: want validation error, got %v", bad, err)
		}
	}
	if _, err := e.user.UpdateProfile(ctx, "missing", models.UserUpdate{Name: sp("x")}); !errors.Is(err, ErrNotFound) {
		t.Fatal("update missing user")
	}
}

func TestUserServiceWithoutCache(t *testing.T) {
	e := newEnv()
	s := NewUserService(e.users, nil, time.Minute)
	u := e.register(t, "n@x.com")
	if got, err := s.Get(context.Background(), u.ID); err != nil || got.ID != u.ID {
		t.Fatal("noop cache path failed")
	}
}

// ---------- tasks ----------

func TestTaskLifecycle(t *testing.T) {
	e := newEnv()
	ctx := context.Background()
	alice := e.register(t, "alice@x.com")
	bob := e.register(t, "bob@x.com")
	carol := e.register(t, "carol@x.com")
	due := time.Now().Add(time.Hour)

	// create with defaults
	d, err := e.task.Create(ctx, alice.ID, CreateTaskInput{Title: "  Write docs "})
	if err != nil {
		t.Fatal(err)
	}
	if d.Title != "Write docs" || d.Status != models.StatusTodo || d.Priority != models.PriorityMedium || d.CreatedBy.ID != alice.ID {
		t.Fatalf("defaults: %+v", d)
	}
	if ev := e.pub.Last(); ev.Type != pubsub.TaskCreated || ev.TaskID != d.ID {
		t.Fatalf("event %+v", ev)
	}

	// create with everything, lower-case enums are normalised
	d2, err := e.task.Create(ctx, alice.ID, CreateTaskInput{Title: "Ship", Description: sp("v1"),
		Status: sp("in_progress"), Priority: sp("high"), AssignedToID: &bob.ID, DueDate: &due})
	if err != nil || d2.Status != models.StatusInProgress || d2.Priority != models.PriorityHigh || d2.AssignedTo.ID != bob.ID {
		t.Fatalf("full create: %+v %v", d2, err)
	}

	// get + cache
	got, err := e.task.Get(ctx, d2.ID)
	if err != nil || got.AssignedTo.Name != bob.Name {
		t.Fatal(err)
	}
	hits := e.cache.Hits
	if _, err := e.task.Get(ctx, d2.ID); err != nil || e.cache.Hits != hits+1 {
		t.Fatal("second Get should hit cache")
	}
	if _, err := e.task.Get(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatal("missing task")
	}

	// list + filters + cache
	all, err := e.task.List(ctx, models.TaskFilter{})
	if err != nil || len(all) != 2 {
		t.Fatalf("list: %d %v", len(all), err)
	}
	hits = e.cache.Hits
	if _, err := e.task.List(ctx, models.TaskFilter{}); err != nil || e.cache.Hits != hits+1 {
		t.Fatal("second List should hit cache")
	}
	inProg, _ := e.task.List(ctx, models.TaskFilter{Status: "in_progress", Priority: "high"})
	if len(inProg) != 1 {
		t.Fatal("filter list")
	}
	if _, err := e.task.List(ctx, models.TaskFilter{Status: "nope"}); !isValidation(err) {
		t.Fatal("bad status filter")
	}
	if _, err := e.task.List(ctx, models.TaskFilter{Priority: "nope"}); !isValidation(err) {
		t.Fatal("bad priority filter")
	}

	// the assignee can move the card
	moved, err := e.task.Update(ctx, bob.ID, d2.ID, models.TaskUpdate{Status: sp("review")})
	if err != nil || moved.Status != models.StatusReview {
		t.Fatalf("assignee move: %v", err)
	}
	if e.pub.Last().Type != pubsub.TaskUpdated {
		t.Fatal("no update event")
	}
	// list cache was invalidated by the update
	reviews, _ := e.task.List(ctx, models.TaskFilter{Status: models.StatusReview})
	if len(reviews) != 1 {
		t.Fatal("stale list after update")
	}
	// strangers cannot edit
	if _, err := e.task.Update(ctx, carol.ID, d2.ID, models.TaskUpdate{Title: sp("x")}); !errors.Is(err, ErrForbidden) {
		t.Fatal("stranger edit allowed")
	}
	// creator edits title
	if u, err := e.task.Update(ctx, alice.ID, d2.ID, models.TaskUpdate{Title: sp(" Ship v2 ")}); err != nil || u.Title != "Ship v2" {
		t.Fatal("creator title edit")
	}
	// assign / unassign
	a, err := e.task.Assign(ctx, alice.ID, d.ID, carol.ID)
	if err != nil || a.AssignedTo.ID != carol.ID {
		t.Fatal("assign")
	}
	ua, err := e.task.Unassign(ctx, alice.ID, d.ID)
	if err != nil || ua.AssignedTo != nil {
		t.Fatal("unassign")
	}
	if u, err := e.task.Update(ctx, alice.ID, d2.ID, models.TaskUpdate{AssignedToID: sp("")}); err != nil || u.AssignedTo != nil {
		t.Fatal("empty assignee should clear")
	}
	if _, err := e.task.Assign(ctx, alice.ID, d.ID, ""); !isValidation(err) {
		t.Fatal("empty assignee id")
	}

	// delete: only creator
	if err := e.task.Delete(ctx, bob.ID, d.ID); !errors.Is(err, ErrForbidden) {
		t.Fatal("non-creator delete allowed")
	}
	if err := e.task.Delete(ctx, alice.ID, d.ID); err != nil {
		t.Fatal(err)
	}
	if ev := e.pub.Last(); ev.Type != pubsub.TaskDeleted || ev.Task != nil {
		t.Fatalf("delete event %+v", ev)
	}
	if err := e.task.Delete(ctx, alice.ID, d.ID); !errors.Is(err, ErrNotFound) {
		t.Fatal("double delete")
	}
	if _, err := e.task.Update(ctx, alice.ID, d.ID, models.TaskUpdate{}); !errors.Is(err, ErrNotFound) {
		t.Fatal("update deleted task")
	}
}

func TestTaskValidation(t *testing.T) {
	e := newEnv()
	ctx := context.Background()
	u := e.register(t, "v@x.com")
	long := make([]byte, 501)
	for i := range long {
		long[i] = 'a'
	}
	bad := []CreateTaskInput{
		{Title: " "},
		{Title: string(long)},
		{Title: "ok", Status: sp("SOMEDAY")},
		{Title: "ok", Priority: sp("CRITICAL")},
		{Title: "ok", AssignedToID: sp("missing-user")},
	}
	for _, in := range bad {
		if _, err := e.task.Create(ctx, u.ID, in); !isValidation(err) {
			t.Errorf("%+v: want validation error, got %v", in, err)
		}
	}
	d, _ := e.task.Create(ctx, u.ID, CreateTaskInput{Title: "ok", AssignedToID: sp("")})
	if d.AssignedTo != nil {
		t.Fatal("empty assignee on create should be ignored")
	}
	for _, upd := range []models.TaskUpdate{
		{Title: sp("")},
		{Status: sp("nope")},
		{Priority: sp("nope")},
		{AssignedToID: sp("missing-user")},
	} {
		if _, err := e.task.Update(ctx, u.ID, d.ID, upd); !isValidation(err) {
			t.Errorf("%+v: want validation error, got %v", upd, err)
		}
	}
}

func TestTaskStoreFailures(t *testing.T) {
	ctx := context.Background()
	setup := func() (*env, *models.User, *models.TaskDetails) {
		e := newEnv()
		u := e.register(t, "f@x.com")
		d, err := e.task.Create(ctx, u.ID, CreateTaskInput{Title: "t"})
		if err != nil {
			t.Fatal(err)
		}
		e.cache.Data = map[string][]byte{} // force store reads
		return e, u, d
	}

	e, u, _ := setup()
	e.tasks.Fail["Create"] = true
	if _, err := e.task.Create(ctx, u.ID, CreateTaskInput{Title: "x"}); !errors.Is(err, memstore.ErrBoom) {
		t.Error("Create failure")
	}
	e, u, _ = setup()
	e.tasks.Fail["CreateRef"] = true
	if _, err := e.task.Create(ctx, u.ID, CreateTaskInput{Title: "x"}); !isValidation(err) {
		t.Error("invalid reference should become a validation error")
	}
	e, u, _ = setup()
	e.users.Fail["GetByIDs"] = true
	if _, err := e.task.Create(ctx, u.ID, CreateTaskInput{Title: "x"}); !errors.Is(err, memstore.ErrBoom) {
		t.Error("hydrate failure on create")
	}
	e, u, _ = setup()
	e.users.Fail["GetByID"] = true
	if _, err := e.task.Create(ctx, u.ID, CreateTaskInput{Title: "x", AssignedToID: sp("someone")}); !errors.Is(err, memstore.ErrBoom) {
		t.Error("assignee lookup failure")
	}

	e, _, d := setup()
	e.users.Fail["GetByIDs"] = true
	if _, err := e.task.Get(ctx, d.ID); !errors.Is(err, memstore.ErrBoom) {
		t.Error("hydrate failure on get")
	}
	if _, err := e.task.List(ctx, models.TaskFilter{}); !errors.Is(err, memstore.ErrBoom) {
		t.Error("hydrate failure on list")
	}
	e, _, _ = setup()
	e.tasks.Fail["List"] = true
	if _, err := e.task.List(ctx, models.TaskFilter{}); !errors.Is(err, memstore.ErrBoom) {
		t.Error("List failure")
	}

	e, u, d = setup()
	e.tasks.Fail["Update"] = true
	if _, err := e.task.Update(ctx, u.ID, d.ID, models.TaskUpdate{Title: sp("y")}); !errors.Is(err, memstore.ErrBoom) {
		t.Error("Update failure")
	}
	e, u, d = setup()
	e.users.Fail["GetByIDs"] = true
	if _, err := e.task.Update(ctx, u.ID, d.ID, models.TaskUpdate{Title: sp("y")}); !errors.Is(err, memstore.ErrBoom) {
		t.Error("hydrate failure on update")
	}
	e, u, d = setup()
	e.tasks.Fail["GetByID"] = true
	if _, err := e.task.Update(ctx, u.ID, d.ID, models.TaskUpdate{}); !errors.Is(err, memstore.ErrBoom) {
		t.Error("GetByID failure on update")
	}
	if err := e.task.Delete(ctx, u.ID, d.ID); !errors.Is(err, memstore.ErrBoom) {
		t.Error("GetByID failure on delete")
	}
	e, u, d = setup()
	e.tasks.Fail["Delete"] = true
	if err := e.task.Delete(ctx, u.ID, d.ID); !errors.Is(err, memstore.ErrBoom) {
		t.Error("Delete failure")
	}
}

func TestTaskServiceDefaults(t *testing.T) {
	e := newEnv()
	s := NewTaskService(e.tasks, e.users, nil, nil, time.Minute)
	u := e.register(t, "d@x.com")
	if _, err := s.Create(context.Background(), u.ID, CreateTaskInput{Title: "x"}); err != nil {
		t.Fatal("noop cache/publisher path failed")
	}
}

func TestFromRepoAndErrors(t *testing.T) {
	if fromRepo(nil) != nil {
		t.Fatal("nil")
	}
	if !errors.Is(fromRepo(repository.ErrNotFound), ErrNotFound) ||
		!errors.Is(fromRepo(repository.ErrDuplicate), ErrConflict) ||
		!isValidation(fromRepo(repository.ErrInvalidReference)) ||
		!errors.Is(fromRepo(memstore.ErrBoom), memstore.ErrBoom) {
		t.Fatal("fromRepo mapping")
	}
	v := &ValidationError{Message: "m"}
	if v.Error() != "m" {
		t.Fatal("ValidationError.Error")
	}
	var nc noopCache
	if nc.Set(context.Background(), "", nil, 0) != nil || nc.Delete(context.Background()) != nil || nc.DeletePattern(context.Background(), "") != nil {
		t.Fatal("noop cache")
	}
}
