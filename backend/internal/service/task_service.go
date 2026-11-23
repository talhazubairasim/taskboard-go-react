package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"taskboard/internal/cache"
	"taskboard/internal/models"
	"taskboard/internal/pubsub"
)

// CreateTaskInput is the data needed to create a task.
type CreateTaskInput struct {
	Title        string     `json:"title"`
	Description  *string    `json:"description"`
	Status       *string    `json:"status"`
	Priority     *string    `json:"priority"`
	AssignedToID *string    `json:"assignedToId"`
	DueDate      *time.Time `json:"dueDate"`
}

// TaskService implements task rules: validation, permissions, caching and events.
//
// Permissions: any signed-in user can see the board and create tasks. The
// creator or the current assignee can edit a task, move it between columns
// and (re)assign it. Only the creator can delete it.
type TaskService struct {
	tasks TaskStore
	users UserStore
	cache Cache
	pub   Publisher
	ttl   time.Duration
}

// NewTaskService builds a TaskService. c and pub may be nil.
func NewTaskService(tasks TaskStore, users UserStore, c Cache, pub Publisher, ttl time.Duration) *TaskService {
	if c == nil {
		c = noopCache{}
	}
	if pub == nil {
		pub = noopPublisher{}
	}
	return &TaskService{tasks: tasks, users: users, cache: c, pub: pub, ttl: ttl}
}

func validateTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" || len(title) > 500 {
		return "", invalid("title must be between 1 and 500 characters")
	}
	return title, nil
}

func normalizeEnum(v *string, valid func(string) bool, field string) (*string, error) {
	if v == nil {
		return nil, nil
	}
	u := strings.ToUpper(strings.TrimSpace(*v))
	if !valid(u) {
		return nil, invalid("invalid %s %q", field, *v)
	}
	return &u, nil
}

func (s *TaskService) checkUserExists(ctx context.Context, id string) error {
	if _, err := s.users.GetByID(ctx, id); err != nil {
		if errors.Is(fromRepo(err), ErrNotFound) {
			return invalid("assignee does not exist")
		}
		return err
	}
	return nil
}

// Create adds a task owned by userID.
func (s *TaskService) Create(ctx context.Context, userID string, in CreateTaskInput) (*models.TaskDetails, error) {
	title, err := validateTitle(in.Title)
	if err != nil {
		return nil, err
	}
	status, err := normalizeEnum(in.Status, models.ValidStatus, "status")
	if err != nil {
		return nil, err
	}
	priority, err := normalizeEnum(in.Priority, models.ValidPriority, "priority")
	if err != nil {
		return nil, err
	}
	t := &models.Task{
		Title:       title,
		Description: in.Description,
		Status:      models.StatusTodo,
		Priority:    models.PriorityMedium,
		CreatedByID: userID,
		DueDate:     in.DueDate,
	}
	if status != nil {
		t.Status = *status
	}
	if priority != nil {
		t.Priority = *priority
	}
	if in.AssignedToID != nil && *in.AssignedToID != "" {
		if err := s.checkUserExists(ctx, *in.AssignedToID); err != nil {
			return nil, err
		}
		t.AssignedToID = in.AssignedToID
	}
	created, err := s.tasks.Create(ctx, t)
	if err != nil {
		return nil, fromRepo(err)
	}
	details, err := s.hydrateOne(ctx, created)
	if err != nil {
		return nil, err
	}
	s.changed(ctx, pubsub.TaskCreated, details.ID, details)
	return details, nil
}

// Get returns one task with its users.
func (s *TaskService) Get(ctx context.Context, id string) (*models.TaskDetails, error) {
	var cached models.TaskDetails
	if err := s.cache.Get(ctx, cache.TaskKey(id), &cached); err == nil {
		return &cached, nil
	}
	t, err := s.tasks.GetByID(ctx, id)
	if err != nil {
		return nil, fromRepo(err)
	}
	details, err := s.hydrateOne(ctx, t)
	if err != nil {
		return nil, err
	}
	_ = s.cache.Set(ctx, cache.TaskKey(id), details, s.ttl)
	return details, nil
}

// List returns the tasks matching f, newest first.
func (s *TaskService) List(ctx context.Context, f models.TaskFilter) ([]*models.TaskDetails, error) {
	if f.Status != "" {
		st, err := normalizeEnum(&f.Status, models.ValidStatus, "status")
		if err != nil {
			return nil, err
		}
		f.Status = *st
	}
	if f.Priority != "" {
		p, err := normalizeEnum(&f.Priority, models.ValidPriority, "priority")
		if err != nil {
			return nil, err
		}
		f.Priority = *p
	}
	key := cache.TasksKey(f.Status + "|" + f.Priority + "|" + f.AssignedToID + "|" + f.CreatedByID)
	var cached []*models.TaskDetails
	if err := s.cache.Get(ctx, key, &cached); err == nil {
		return cached, nil
	}
	tasks, err := s.tasks.List(ctx, f)
	if err != nil {
		return nil, fromRepo(err)
	}
	out, err := s.hydrate(ctx, tasks)
	if err != nil {
		return nil, err
	}
	_ = s.cache.Set(ctx, key, out, s.ttl)
	return out, nil
}

// Update edits a task. Only the creator or the assignee may do this.
func (s *TaskService) Update(ctx context.Context, userID, id string, u models.TaskUpdate) (*models.TaskDetails, error) {
	existing, err := s.tasks.GetByID(ctx, id)
	if err != nil {
		return nil, fromRepo(err)
	}
	if !canEdit(existing, userID) {
		return nil, ErrForbidden
	}
	if u.Title != nil {
		title, err := validateTitle(*u.Title)
		if err != nil {
			return nil, err
		}
		u.Title = &title
	}
	if u.Status, err = normalizeEnum(u.Status, models.ValidStatus, "status"); err != nil {
		return nil, err
	}
	if u.Priority, err = normalizeEnum(u.Priority, models.ValidPriority, "priority"); err != nil {
		return nil, err
	}
	if u.AssignedToID != nil {
		if *u.AssignedToID == "" {
			u.AssignedToID, u.ClearAssignee = nil, true
		} else if err := s.checkUserExists(ctx, *u.AssignedToID); err != nil {
			return nil, err
		}
	}
	updated, err := s.tasks.Update(ctx, id, u)
	if err != nil {
		return nil, fromRepo(err)
	}
	details, err := s.hydrateOne(ctx, updated)
	if err != nil {
		return nil, err
	}
	s.changed(ctx, pubsub.TaskUpdated, id, details)
	return details, nil
}

// Assign sets the assignee.
func (s *TaskService) Assign(ctx context.Context, userID, taskID, assigneeID string) (*models.TaskDetails, error) {
	if assigneeID == "" {
		return nil, invalid("userId is required")
	}
	return s.Update(ctx, userID, taskID, models.TaskUpdate{AssignedToID: &assigneeID})
}

// Unassign clears the assignee.
func (s *TaskService) Unassign(ctx context.Context, userID, taskID string) (*models.TaskDetails, error) {
	return s.Update(ctx, userID, taskID, models.TaskUpdate{ClearAssignee: true})
}

// Delete removes a task. Only its creator may do this.
func (s *TaskService) Delete(ctx context.Context, userID, id string) error {
	existing, err := s.tasks.GetByID(ctx, id)
	if err != nil {
		return fromRepo(err)
	}
	if existing.CreatedByID != userID {
		return ErrForbidden
	}
	if err := s.tasks.Delete(ctx, id); err != nil {
		return fromRepo(err)
	}
	s.changed(ctx, pubsub.TaskDeleted, id, nil)
	return nil
}

func canEdit(t *models.Task, userID string) bool {
	return t.CreatedByID == userID || (t.AssignedToID != nil && *t.AssignedToID == userID)
}

// changed invalidates the cache and notifies subscribers.
func (s *TaskService) changed(ctx context.Context, typ pubsub.EventType, id string, t *models.TaskDetails) {
	_ = s.cache.Delete(ctx, cache.TaskKey(id))
	_ = s.cache.DeletePattern(ctx, cache.TasksPrefix+"*")
	s.pub.Publish(pubsub.Event{Type: typ, TaskID: id, Task: t})
}

func (s *TaskService) hydrateOne(ctx context.Context, t *models.Task) (*models.TaskDetails, error) {
	out, err := s.hydrate(ctx, []*models.Task{t})
	if err != nil {
		return nil, err
	}
	return out[0], nil
}

// hydrate loads the creator and assignee of every task with one query.
func (s *TaskService) hydrate(ctx context.Context, tasks []*models.Task) ([]*models.TaskDetails, error) {
	seen := map[string]bool{}
	var ids []string
	for _, t := range tasks {
		for _, id := range []*string{&t.CreatedByID, t.AssignedToID} {
			if id != nil && !seen[*id] {
				seen[*id] = true
				ids = append(ids, *id)
			}
		}
	}
	users, err := s.users.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make([]*models.TaskDetails, 0, len(tasks))
	for _, t := range tasks {
		d := &models.TaskDetails{Task: *t, CreatedBy: users[t.CreatedByID]}
		if t.AssignedToID != nil {
			d.AssignedTo = users[*t.AssignedToID]
		}
		out = append(out, d)
	}
	return out, nil
}
