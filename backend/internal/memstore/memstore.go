// Package memstore provides in-memory implementations of the service
// interfaces. Tests use them so the business rules can be exercised without a
// database; any method can be made to fail through the Fail map.
package memstore

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"taskboard/internal/cache"
	"taskboard/internal/models"
	"taskboard/internal/pubsub"
	"taskboard/internal/repository"
)

// ErrBoom is returned by any method whose name is set in a Fail map.
var ErrBoom = errors.New("boom")

// Users is an in-memory UserStore. Set fail to make a method return ErrBoom.
type Users struct {
	mu   sync.Mutex
	ByID map[string]*models.User
	Fail map[string]bool
}

// NewUsers returns an empty user store.
func NewUsers() *Users {
	return &Users{ByID: map[string]*models.User{}, Fail: map[string]bool{}}
}

func (f *Users) Create(_ context.Context, u *models.User) (*models.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Fail["Create"] {
		return nil, ErrBoom
	}
	for _, x := range f.ByID {
		if x.Email == u.Email {
			return nil, repository.ErrDuplicate
		}
	}
	u.ID = uuid.NewString()
	u.CreatedAt, u.UpdatedAt = time.Now(), time.Now()
	c := *u
	f.ByID[u.ID] = &c
	return u, nil
}

func (f *Users) GetByID(_ context.Context, id string) (*models.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Fail["GetByID"] {
		return nil, ErrBoom
	}
	u, ok := f.ByID[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	c := *u
	return &c, nil
}

func (f *Users) GetByEmail(_ context.Context, email string) (*models.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Fail["GetByEmail"] {
		return nil, ErrBoom
	}
	for _, u := range f.ByID {
		if u.Email == strings.ToLower(email) {
			c := *u
			return &c, nil
		}
	}
	return nil, repository.ErrNotFound
}

func (f *Users) GetByIDs(_ context.Context, ids []string) (map[string]*models.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Fail["GetByIDs"] {
		return nil, ErrBoom
	}
	out := map[string]*models.User{}
	for _, id := range ids {
		if u, ok := f.ByID[id]; ok {
			c := *u
			out[id] = &c
		}
	}
	return out, nil
}

func (f *Users) List(context.Context) ([]*models.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Fail["List"] {
		return nil, ErrBoom
	}
	out := []*models.User{}
	for _, u := range f.ByID {
		c := *u
		out = append(out, &c)
	}
	return out, nil
}

func (f *Users) Update(_ context.Context, id string, upd models.UserUpdate) (*models.User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Fail["Update"] {
		return nil, ErrBoom
	}
	u, ok := f.ByID[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	if upd.Name != nil {
		u.Name = *upd.Name
	}
	if upd.Avatar != nil {
		a := *upd.Avatar
		u.Avatar = &a
	}
	c := *u
	return &c, nil
}

func (f *Users) EmailExists(_ context.Context, email string) (bool, error) {
	if f.Fail["EmailExists"] {
		return false, ErrBoom
	}
	_, err := f.GetByEmail(context.Background(), email)
	return err == nil, nil
}

// Tasks is an in-memory TaskStore.
type Tasks struct {
	mu   sync.Mutex
	ByID map[string]*models.Task
	Fail map[string]bool
	seq  int
}

// NewTasks returns an empty task store.
func NewTasks() *Tasks {
	return &Tasks{ByID: map[string]*models.Task{}, Fail: map[string]bool{}}
}

func (f *Tasks) Create(_ context.Context, t *models.Task) (*models.Task, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Fail["Create"] {
		return nil, ErrBoom
	}
	if f.Fail["CreateRef"] {
		return nil, repository.ErrInvalidReference
	}
	f.seq++
	t.ID = uuid.NewString()
	t.CreatedAt = time.Unix(int64(f.seq), 0)
	t.UpdatedAt = t.CreatedAt
	c := *t
	f.ByID[t.ID] = &c
	return t, nil
}

func (f *Tasks) GetByID(_ context.Context, id string) (*models.Task, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Fail["GetByID"] {
		return nil, ErrBoom
	}
	t, ok := f.ByID[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	c := *t
	return &c, nil
}

func (f *Tasks) List(_ context.Context, flt models.TaskFilter) ([]*models.Task, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Fail["List"] {
		return nil, ErrBoom
	}
	out := []*models.Task{}
	for _, t := range f.ByID {
		if flt.Status != "" && t.Status != flt.Status ||
			flt.Priority != "" && t.Priority != flt.Priority ||
			flt.CreatedByID != "" && t.CreatedByID != flt.CreatedByID ||
			flt.AssignedToID != "" && (t.AssignedToID == nil || *t.AssignedToID != flt.AssignedToID) {
			continue
		}
		c := *t
		out = append(out, &c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (f *Tasks) Update(_ context.Context, id string, u models.TaskUpdate) (*models.Task, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Fail["Update"] {
		return nil, ErrBoom
	}
	t, ok := f.ByID[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	if u.Title != nil {
		t.Title = *u.Title
	}
	if u.ClearDescription {
		t.Description = nil
	} else if u.Description != nil {
		t.Description = u.Description
	}
	if u.Status != nil {
		t.Status = *u.Status
	}
	if u.Priority != nil {
		t.Priority = *u.Priority
	}
	if u.ClearAssignee {
		t.AssignedToID = nil
	} else if u.AssignedToID != nil {
		a := *u.AssignedToID
		t.AssignedToID = &a
	}
	if u.ClearDueDate {
		t.DueDate = nil
	} else if u.DueDate != nil {
		t.DueDate = u.DueDate
	}
	c := *t
	return &c, nil
}

func (f *Tasks) Delete(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Fail["Delete"] {
		return ErrBoom
	}
	if _, ok := f.ByID[id]; !ok {
		return repository.ErrNotFound
	}
	delete(f.ByID, id)
	return nil
}

// Cache is an in-memory Cache that records hits.
type Cache struct {
	mu   sync.Mutex
	Data map[string][]byte
	Hits int
}

// NewCache returns an empty cache.
func NewCache() *Cache { return &Cache{Data: map[string][]byte{}} }

func (c *Cache) Get(_ context.Context, key string, dest interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	b, ok := c.Data[key]
	if !ok {
		return cache.ErrCacheMiss
	}
	c.Hits++
	return json.Unmarshal(b, dest)
}

func (c *Cache) Set(_ context.Context, key string, v interface{}, _ time.Duration) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.Data[key] = b
	c.mu.Unlock()
	return nil
}

func (c *Cache) Delete(_ context.Context, keys ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, k := range keys {
		delete(c.Data, k)
	}
	return nil
}

func (c *Cache) DeletePattern(_ context.Context, pattern string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	prefix := strings.TrimSuffix(pattern, "*")
	for k := range c.Data {
		if strings.HasPrefix(k, prefix) {
			delete(c.Data, k)
		}
	}
	return nil
}

// Recorder collects published events.
type Recorder struct {
	mu     sync.Mutex
	Events []pubsub.Event
}

func (r *Recorder) Publish(e pubsub.Event) {
	r.mu.Lock()
	r.Events = append(r.Events, e)
	r.mu.Unlock()
}

// Last returns the most recent event.
func (r *Recorder) Last() pubsub.Event {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Events[len(r.Events)-1]
}
