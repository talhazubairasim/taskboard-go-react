package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"

	"taskboard/internal/auth"
	"taskboard/internal/cache"
	"taskboard/internal/memstore"
	"taskboard/internal/metrics"
	"taskboard/internal/pubsub"
	"taskboard/internal/service"
)

func init() { auth.BcryptCost = bcrypt.MinCost }

const pw = "Str0ng!Pass"

type harness struct {
	t     *testing.T
	srv   *httptest.Server
	users *memstore.Users
	tasks *memstore.Tasks
	jwt   *auth.JWTManager
	ready error
}

func newHarness(t *testing.T, env string) *harness {
	t.Helper()
	h := &harness{t: t, users: memstore.NewUsers(), tasks: memstore.NewTasks(), jwt: auth.NewJWTManager("s", "r")}
	c := memstore.NewCache()
	broker := pubsub.NewBroker()
	db := metrics.NewRecorder(16)
	db.Observe(time.Millisecond)
	h.srv = httptest.NewServer(New(Deps{
		Env:            env,
		AllowedOrigins: []string{"http://allowed.test"},
		JWT:            h.jwt,
		AuthSvc:        service.NewAuthService(h.users, h.jwt),
		UserSvc:        service.NewUserService(h.users, c, time.Minute),
		TaskSvc:        service.NewTaskService(h.tasks, h.users, c, broker, time.Minute),
		Broker:         broker,
		Ready:          func(context.Context) error { return h.ready },
		CacheStats:     &cache.Stats{},
		DBLatency:      db,
	}))
	t.Cleanup(h.srv.Close)
	return h
}

// do sends a JSON request and decodes a JSON response into out (if non-nil).
func (h *harness) do(method, path, token string, body interface{}, out interface{}) int {
	h.t.Helper()
	var r io.Reader
	if s, ok := body.(string); ok {
		r = strings.NewReader(s)
	} else if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewReader(b)
	}
	req, _ := http.NewRequest(method, h.srv.URL+path, r)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		h.t.Fatal(err)
	}
	defer res.Body.Close()
	if out != nil {
		_ = json.NewDecoder(res.Body).Decode(out)
	}
	return res.StatusCode
}

type gqlResp struct {
	Data   map[string]json.RawMessage `json:"data"`
	Errors []struct {
		Message    string                 `json:"message"`
		Extensions map[string]interface{} `json:"extensions"`
	} `json:"errors"`
}

func (g gqlResp) code() string {
	if len(g.Errors) == 0 {
		return ""
	}
	c, _ := g.Errors[0].Extensions["code"].(string)
	return c
}

func (h *harness) gql(token, query string, vars map[string]interface{}) gqlResp {
	h.t.Helper()
	var out gqlResp
	h.do("POST", "/query", token, map[string]interface{}{"query": query, "variables": vars}, &out)
	return out
}

type authResp struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	User         struct {
		ID string `json:"id"`
	} `json:"user"`
}

func (h *harness) register(email string) authResp {
	h.t.Helper()
	var a authResp
	if code := h.do("POST", "/api/auth/register", "", map[string]string{"email": email, "password": pw, "name": email}, &a); code != 201 {
		h.t.Fatalf("register %s: %d", email, code)
	}
	return a
}

type taskJSON struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Status      string  `json:"status"`
	Description *string `json:"description"`
	DueDate     *string `json:"dueDate"`
	CreatedBy   *struct {
		ID string `json:"id"`
	} `json:"createdBy"`
	AssignedTo *struct {
		ID string `json:"id"`
	} `json:"assignedTo"`
}

func TestRESTFlow(t *testing.T) {
	h := newHarness(t, "development")
	alice := h.register("alice@x.com")
	bob := h.register("bob@x.com")

	// auth endpoints
	var a authResp
	if code := h.do("POST", "/api/auth/login", "", map[string]string{"email": "alice@x.com", "password": pw}, &a); code != 200 || a.Token == "" {
		t.Fatalf("login %d", code)
	}
	if code := h.do("POST", "/api/auth/login", "", map[string]string{"email": "alice@x.com", "password": "nope"}, nil); code != 401 {
		t.Fatalf("bad login %d", code)
	}
	if code := h.do("POST", "/api/auth/refresh", "", map[string]string{"refreshToken": a.RefreshToken}, &a); code != 200 {
		t.Fatalf("refresh %d", code)
	}
	if code := h.do("POST", "/api/auth/refresh", "", map[string]string{"refreshToken": "junk"}, nil); code != 401 {
		t.Fatalf("bad refresh %d", code)
	}
	if code := h.do("POST", "/api/auth/register", "", map[string]string{"email": "alice@x.com", "password": pw, "name": "A"}, nil); code != 409 {
		t.Fatalf("duplicate register %d", code)
	}
	for _, p := range []string{"/api/auth/register", "/api/auth/login", "/api/auth/refresh"} {
		if code := h.do("POST", p, "", "{not json", nil); code != 400 {
			t.Fatalf("%s bad json %d", p, code)
		}
	}
	if code := h.do("POST", "/api/auth/register", "", map[string]string{"email": "x", "password": pw, "name": "A"}, nil); code != 400 {
		t.Fatalf("invalid register %d", code)
	}

	// anonymous and bad tokens
	if code := h.do("GET", "/api/tasks", "", nil, nil); code != 401 {
		t.Fatalf("anonymous %d", code)
	}
	if code := h.do("GET", "/api/tasks", "garbage", nil, nil); code != 401 {
		t.Fatalf("bad token %d", code)
	}

	// users
	var me struct {
		ID, Name string
		Avatar   *string
	}
	if code := h.do("GET", "/api/me", alice.Token, nil, &me); code != 200 || me.ID != alice.User.ID {
		t.Fatalf("me %d", code)
	}
	if code := h.do("PATCH", "/api/me", alice.Token, map[string]string{"name": "Alice", "avatar": "https://a/b.png"}, &me); code != 200 || me.Name != "Alice" {
		t.Fatalf("patch me %d", code)
	}
	if code := h.do("PATCH", "/api/me", alice.Token, map[string]string{"avatar": "ftp://x"}, nil); code != 400 {
		t.Fatalf("bad avatar %d", code)
	}
	if code := h.do("PATCH", "/api/me", alice.Token, "[", nil); code != 400 {
		t.Fatalf("bad json %d", code)
	}
	var users []map[string]interface{}
	if code := h.do("GET", "/api/users", alice.Token, nil, &users); code != 200 || len(users) != 2 {
		t.Fatalf("users %d %d", code, len(users))
	}
	if _, leaked := users[0]["passwordHash"]; leaked {
		t.Fatal("password hash leaked")
	}
	if code := h.do("GET", "/api/users/"+bob.User.ID, alice.Token, nil, nil); code != 200 {
		t.Fatalf("get user %d", code)
	}
	if code := h.do("GET", "/api/users/missing", alice.Token, nil, nil); code != 404 {
		t.Fatalf("missing user %d", code)
	}

	// tasks
	var task taskJSON
	if code := h.do("POST", "/api/tasks", alice.Token, map[string]interface{}{
		"title": "Design", "description": "d", "priority": "HIGH", "assignedToId": bob.User.ID, "dueDate": "2030-01-02T15:04:05Z",
	}, &task); code != 201 || task.AssignedTo.ID != bob.User.ID || task.CreatedBy.ID != alice.User.ID {
		t.Fatalf("create %d %+v", code, task)
	}
	if code := h.do("POST", "/api/tasks", alice.Token, map[string]string{"title": ""}, nil); code != 400 {
		t.Fatalf("empty title %d", code)
	}
	if code := h.do("POST", "/api/tasks", alice.Token, "x", nil); code != 400 {
		t.Fatalf("bad json %d", code)
	}
	var list []taskJSON
	if code := h.do("GET", "/api/tasks?status=TODO&priority=HIGH", alice.Token, nil, &list); code != 200 || len(list) != 1 {
		t.Fatalf("list %d %d", code, len(list))
	}
	if code := h.do("GET", "/api/tasks?status=bogus", alice.Token, nil, nil); code != 400 {
		t.Fatalf("bad filter %d", code)
	}
	if code := h.do("GET", "/api/tasks/mine", alice.Token, nil, &list); code != 200 || len(list) != 1 {
		t.Fatalf("mine %d", code)
	}
	if code := h.do("GET", "/api/tasks/assigned", bob.Token, nil, &list); code != 200 || len(list) != 1 {
		t.Fatalf("assigned %d", code)
	}
	if code := h.do("GET", "/api/tasks/"+task.ID, bob.Token, nil, &task); code != 200 {
		t.Fatalf("get %d", code)
	}
	if code := h.do("GET", "/api/tasks/missing", bob.Token, nil, nil); code != 404 {
		t.Fatalf("get missing %d", code)
	}

	// update: assignee moves the card; nulls clear fields
	if code := h.do("PATCH", "/api/tasks/"+task.ID, bob.Token, map[string]interface{}{"status": "IN_PROGRESS", "title": "Design v2", "priority": "LOW"}, &task); code != 200 || task.Status != "IN_PROGRESS" {
		t.Fatalf("patch %d", code)
	}
	if code := h.do("PUT", "/api/tasks/"+task.ID, alice.Token, `{"description":null,"dueDate":null}`, &task); code != 200 || task.Description != nil || task.DueDate != nil {
		t.Fatalf("clear fields %d %+v", code, task)
	}
	if code := h.do("PATCH", "/api/tasks/"+task.ID, alice.Token, map[string]string{"dueDate": "2031-05-06T00:00:00Z", "description": "again"}, &task); code != 200 || task.DueDate == nil {
		t.Fatalf("set due date %d", code)
	}
	for _, bad := range []string{`{"dueDate":"tomorrow"}`, `{"title":5}`, `{"description":1}`, `{"status":1}`, `{"priority":1}`, `{"assignedToId":1}`, `{"dueDate":1}`, `nope`} {
		if code := h.do("PATCH", "/api/tasks/"+task.ID, alice.Token, bad, nil); code != 400 {
			t.Fatalf("%s: %d", bad, code)
		}
	}
	carol := h.register("carol@x.com")
	if code := h.do("PATCH", "/api/tasks/"+task.ID, carol.Token, map[string]string{"title": "hijack"}, nil); code != 403 {
		t.Fatalf("stranger edit %d", code)
	}

	// assign / unassign
	if code := h.do("POST", "/api/tasks/"+task.ID+"/assign", alice.Token, map[string]string{"userId": carol.User.ID}, &task); code != 200 || task.AssignedTo.ID != carol.User.ID {
		t.Fatalf("assign %d", code)
	}
	if code := h.do("POST", "/api/tasks/"+task.ID+"/assign", alice.Token, "{", nil); code != 400 {
		t.Fatalf("assign bad json %d", code)
	}
	if code := h.do("POST", "/api/tasks/"+task.ID+"/assign", alice.Token, map[string]string{"userId": "nobody"}, nil); code != 400 {
		t.Fatalf("assign missing user %d", code)
	}
	if code := h.do("POST", "/api/tasks/"+task.ID+"/unassign", alice.Token, nil, &task); code != 200 || task.AssignedTo != nil {
		t.Fatalf("unassign %d", code)
	}
	if code := h.do("POST", "/api/tasks/missing/unassign", alice.Token, nil, nil); code != 404 {
		t.Fatalf("unassign missing %d", code)
	}

	// delete
	if code := h.do("DELETE", "/api/tasks/"+task.ID, bob.Token, nil, nil); code != 403 {
		t.Fatalf("non-creator delete %d", code)
	}
	if code := h.do("DELETE", "/api/tasks/"+task.ID, alice.Token, nil, nil); code != 204 {
		t.Fatalf("delete %d", code)
	}

	// internal errors are hidden
	h.tasks.Fail["List"] = true
	var e map[string]string
	if code := h.do("GET", "/api/tasks?priority=URGENT", alice.Token, nil, &e); code != 500 || e["error"] != "internal server error" {
		t.Fatalf("internal error %d %v", code, e)
	}
	h.users.Fail["List"] = true
	if code := h.do("GET", "/api/users", alice.Token, nil, nil); code != 500 {
		t.Fatalf("users internal %d", code)
	}
	h.users.Fail["Update"] = true
	if code := h.do("PATCH", "/api/me", alice.Token, map[string]string{"name": "x"}, nil); code != 500 {
		t.Fatalf("patch me internal %d", code)
	}
	h.users.Fail["GetByID"] = true
	if code := h.do("GET", "/api/me", h.register2(t, "dave@x.com"), nil, nil); code != 500 {
		t.Fatalf("me internal %d", code)
	}
}

// register2 registers without going through users.GetByID (which may be failing).
func (h *harness) register2(t *testing.T, email string) string {
	tok, _ := h.jwt.GenerateAccessToken("ghost-"+email, email)
	return tok
}

func TestGraphQLFlow(t *testing.T) {
	h := newHarness(t, "development")

	reg := h.gql("", `mutation($i: RegisterInput!){ register(input:$i){ token refreshToken user{ id email name } } }`,
		map[string]interface{}{"i": map[string]string{"email": "g@x.com", "password": pw, "name": "G"}})
	if len(reg.Errors) > 0 {
		t.Fatalf("register: %+v", reg.Errors)
	}
	var payload authResp
	_ = json.Unmarshal(reg.Data["register"], &payload)
	tok := payload.Token
	other := h.register("o@x.com")

	if r := h.gql("", `mutation{ login(input:{email:"g@x.com", password:"`+pw+`"}){ token } }`, nil); len(r.Errors) > 0 {
		t.Fatalf("login: %+v", r.Errors)
	}
	if r := h.gql("", `mutation{ login(input:{email:"g@x.com", password:"bad"}){ token } }`, nil); r.code() != "UNAUTHENTICATED" {
		t.Fatalf("bad login code %q", r.code())
	}
	if r := h.gql("", `mutation($t:String!){ refreshToken(refreshToken:$t){ token } }`, map[string]interface{}{"t": payload.RefreshToken}); len(r.Errors) > 0 {
		t.Fatalf("refresh: %+v", r.Errors)
	}
	if r := h.gql("", `mutation{ refreshToken(refreshToken:"x"){ token } }`, nil); r.code() != "UNAUTHENTICATED" {
		t.Fatal("bad refresh")
	}
	if r := h.gql("", `mutation{ register(input:{email:"bad", password:"x", name:""}){ token } }`, nil); r.code() != "BAD_USER_INPUT" {
		t.Fatalf("invalid register code %q", r.code())
	}

	// every protected field rejects anonymous callers with UNAUTHENTICATED
	for _, q := range []string{
		`{ me { id } }`, `{ users { id } }`, `{ user(id:"x"){ id } }`, `{ tasks { id } }`, `{ task(id:"x"){ id } }`,
		`{ myTasks { id } }`, `{ assignedTasks { id } }`,
		`mutation{ createTask(input:{title:"x"}){ id } }`, `mutation{ updateTask(id:"x", input:{}){ id } }`,
		`mutation{ deleteTask(id:"x") }`, `mutation{ assignTask(taskId:"x", userId:"y"){ id } }`,
		`mutation{ unassignTask(taskId:"x"){ id } }`, `mutation{ updateProfile(name:"x"){ id } }`,
	} {
		if r := h.gql("", q, nil); r.code() != "UNAUTHENTICATED" {
			t.Errorf("%s: code %q", q, r.code())
		}
	}

	if r := h.gql(tok, `{ me { id email } users { id } user(id:"`+other.User.ID+`"){ name } }`, nil); len(r.Errors) > 0 {
		t.Fatalf("user queries: %+v", r.Errors)
	}
	if r := h.gql(tok, `{ user(id:"missing"){ id } }`, nil); r.code() != "NOT_FOUND" {
		t.Fatalf("missing user code %q", r.code())
	}
	if r := h.gql(tok, `mutation{ updateProfile(name:"Gee", avatar:"https://x/y.png"){ name avatar } }`, nil); len(r.Errors) > 0 {
		t.Fatalf("updateProfile: %+v", r.Errors)
	}
	if r := h.gql(tok, `mutation{ updateProfile(name:""){ name } }`, nil); r.code() != "BAD_USER_INPUT" {
		t.Fatal("invalid profile")
	}

	cr := h.gql(tok, `mutation($i: CreateTaskInput!){ createTask(input:$i){ id title status priority dueDate createdBy{ id } assignedTo{ id } } }`,
		map[string]interface{}{"i": map[string]interface{}{"title": "GQL task", "status": "REVIEW", "priority": "URGENT",
			"assignedToId": other.User.ID, "dueDate": "2030-01-01T00:00:00Z", "description": "d"}})
	if len(cr.Errors) > 0 {
		t.Fatalf("createTask: %+v", cr.Errors)
	}
	var created taskJSON
	_ = json.Unmarshal(cr.Data["createTask"], &created)
	if created.Status != "REVIEW" || created.AssignedTo.ID != other.User.ID {
		t.Fatalf("created %+v", created)
	}
	if r := h.gql(tok, `mutation{ createTask(input:{title:"  "}){ id } }`, nil); r.code() != "BAD_USER_INPUT" {
		t.Fatal("blank title")
	}
	if r := h.gql(tok, `mutation{ createTask(input:{title:"x", status: NOPE}){ id } }`, nil); r.code() == "" {
		t.Fatal("bad enum should fail validation")
	}

	q := `{ tasks(filter:{status: REVIEW, priority: URGENT, assignedToId:"` + other.User.ID + `", createdById:"` + payload.User.ID + `"}){ id }
		all: tasks { id } todo: tasks(filter:{status: TODO}){ id } task(id:"` + created.ID + `"){ title } myTasks { id } assignedTasks { id } }`
	if r := h.gql(tok, q, nil); len(r.Errors) > 0 {
		t.Fatalf("task queries: %+v", r.Errors)
	}
	if r := h.gql(tok, `{ task(id:"missing"){ id } }`, nil); r.code() != "NOT_FOUND" {
		t.Fatal("missing task")
	}

	up := h.gql(other.Token, `mutation{ updateTask(id:"`+created.ID+`", input:{status: DONE, title:"Done", priority: LOW}){ status } }`, nil)
	if len(up.Errors) > 0 {
		t.Fatalf("updateTask: %+v", up.Errors)
	}
	if r := h.gql(other.Token, `mutation{ updateTask(id:"`+created.ID+`", input:{status: DONE}){ id } }`, nil); len(r.Errors) > 0 {
		t.Fatal("assignee should be allowed to update")
	}
	if r := h.gql(tok, `mutation{ assignTask(taskId:"`+created.ID+`", userId:"`+payload.User.ID+`"){ assignedTo{ id } } }`, nil); len(r.Errors) > 0 {
		t.Fatalf("assignTask: %+v", r.Errors)
	}
	if r := h.gql(tok, `mutation{ assignTask(taskId:"`+created.ID+`", userId:"missing"){ id } }`, nil); r.code() != "BAD_USER_INPUT" {
		t.Fatal("assign missing user")
	}
	if r := h.gql(other.Token, `mutation{ updateTask(id:"`+created.ID+`", input:{title:"x"}){ id } }`, nil); r.code() != "FORBIDDEN" {
		t.Fatalf("forbidden code %q", r.code())
	}
	if r := h.gql(tok, `mutation{ unassignTask(taskId:"`+created.ID+`"){ assignedTo{ id } } }`, nil); len(r.Errors) > 0 {
		t.Fatalf("unassignTask: %+v", r.Errors)
	}
	if r := h.gql(tok, `mutation{ unassignTask(taskId:"missing"){ id } }`, nil); r.code() != "NOT_FOUND" {
		t.Fatal("unassign missing")
	}
	if r := h.gql(other.Token, `mutation{ deleteTask(id:"`+created.ID+`") }`, nil); r.code() != "FORBIDDEN" {
		t.Fatal("non-creator delete")
	}
	if r := h.gql(tok, `mutation{ deleteTask(id:"`+created.ID+`") }`, nil); len(r.Errors) > 0 {
		t.Fatalf("deleteTask: %+v", r.Errors)
	}

	// internal errors are masked, everything else keeps its message
	h.tasks.Fail["List"] = true
	h.users.Fail["List"] = true
	for _, q := range []string{`{ tasks { id } }`, `{ myTasks { id } }`, `{ assignedTasks { id } }`, `{ users { id } }`} {
		r := h.gql(tok, q, nil)
		if r.code() != "INTERNAL_SERVER_ERROR" || r.Errors[0].Message != "internal server error" {
			t.Errorf("%s: %+v", q, r.Errors)
		}
	}
	h.tasks.Fail["Create"] = true
	if r := h.gql(tok, `mutation{ createTask(input:{title:"x"}){ id } }`, nil); r.code() != "INTERNAL_SERVER_ERROR" {
		t.Fatal("create internal")
	}
	h.users.Fail["Update"] = true
	if r := h.gql(tok, `mutation{ updateProfile(name:"x"){ id } }`, nil); r.code() != "INTERNAL_SERVER_ERROR" {
		t.Fatal("updateProfile internal")
	}
	h.users.Fail["GetByEmail"] = true
	if r := h.gql("", `mutation{ login(input:{email:"g@x.com", password:"x"}){ token } }`, nil); r.code() != "INTERNAL_SERVER_ERROR" {
		t.Fatal("login internal")
	}
	h.users.Fail["GetByID"] = true
	if r := h.gql(tok, `{ me { id } }`, nil); r.code() != "INTERNAL_SERVER_ERROR" {
		t.Fatal("me internal")
	}
	if r := h.gql("", `mutation($t:String!){ refreshToken(refreshToken:$t){ token } }`, map[string]interface{}{"t": payload.RefreshToken}); r.code() != "INTERNAL_SERVER_ERROR" {
		t.Fatal("refresh internal")
	}
	h.users.Fail["EmailExists"] = true
	if r := h.gql("", `mutation{ register(input:{email:"z@x.com", password:"`+pw+`", name:"Z"}){ token } }`, nil); r.code() != "INTERNAL_SERVER_ERROR" {
		t.Fatal("register internal")
	}
	h.tasks.Fail["GetByID"] = true
	for _, q := range []string{
		`mutation{ updateTask(id:"x", input:{}){ id } }`, `mutation{ deleteTask(id:"x") }`,
		`mutation{ assignTask(taskId:"x", userId:"y"){ id } }`,
	} {
		if r := h.gql(tok, q, nil); r.code() != "INTERNAL_SERVER_ERROR" {
			t.Errorf("%s: %q", q, r.code())
		}
	}

	// gqlgen's own errors keep a code too
	if r := h.gql(tok, `{ nope }`, nil); r.code() == "" {
		t.Fatal("validation error without code")
	}
	if r := h.gql(tok, `query($id: ID!){ task(id:$id){ id } }`, map[string]interface{}{}); r.code() == "" {
		t.Fatal("missing variable error without code")
	}
}

// ---------- WebSocket subscriptions ----------

type wsMsg struct {
	ID      string          `json:"id,omitempty"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

func dialWS(t *testing.T, h *harness, token string, origin string) (*websocket.Conn, error) {
	t.Helper()
	url := "ws" + strings.TrimPrefix(h.srv.URL, "http") + "/query"
	hdr := http.Header{}
	if origin != "" {
		hdr.Set("Origin", origin)
	}
	d := websocket.Dialer{Subprotocols: []string{"graphql-transport-ws"}}
	c, _, err := d.Dial(url, hdr)
	if err != nil {
		return nil, err
	}
	init := map[string]interface{}{"type": "connection_init", "payload": map[string]string{}}
	if token != "" {
		init["payload"] = map[string]string{"Authorization": "Bearer " + token}
	}
	if err := c.WriteJSON(init); err != nil {
		return nil, err
	}
	var ack wsMsg
	_ = c.SetReadDeadline(time.Now().Add(3 * time.Second))
	if err := c.ReadJSON(&ack); err != nil {
		c.Close()
		return nil, err
	}
	if ack.Type != "connection_ack" {
		c.Close()
		return nil, errors.New("no ack: " + ack.Type)
	}
	return c, nil
}

func subscribe(t *testing.T, c *websocket.Conn, id, query string) {
	t.Helper()
	if err := c.WriteJSON(map[string]interface{}{"id": id, "type": "subscribe", "payload": map[string]string{"query": query}}); err != nil {
		t.Fatal(err)
	}
}

func next(t *testing.T, c *websocket.Conn) wsMsg {
	t.Helper()
	for {
		var m wsMsg
		_ = c.SetReadDeadline(time.Now().Add(3 * time.Second))
		if err := c.ReadJSON(&m); err != nil {
			t.Fatalf("read: %v", err)
		}
		if m.Type == "ping" || m.Type == "pong" || m.Type == "ka" || m.Type == "complete" {
			continue
		}
		return m
	}
}

func TestSubscriptions(t *testing.T) {
	h := newHarness(t, "development")
	a := h.register("ws@x.com")

	c, err := dialWS(t, h, a.Token, "")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	subscribe(t, c, "1", `subscription{ taskCreated{ id title } }`)
	time.Sleep(100 * time.Millisecond)

	var task taskJSON
	h.do("POST", "/api/tasks", a.Token, map[string]string{"title": "live"}, &task)
	m := next(t, c)
	if m.Type != "next" || !strings.Contains(string(m.Payload), "live") {
		t.Fatalf("taskCreated: %+v %s", m, m.Payload)
	}

	subscribe(t, c, "2", `subscription{ taskUpdated(taskId:"`+task.ID+`"){ id status } }`)
	subscribe(t, c, "3", `subscription{ taskDeleted }`)
	time.Sleep(100 * time.Millisecond)
	h.do("PATCH", "/api/tasks/"+task.ID, a.Token, map[string]string{"status": "DONE"}, nil)
	m = next(t, c)
	if m.ID != "2" || !strings.Contains(string(m.Payload), "DONE") {
		t.Fatalf("taskUpdated: %+v %s", m, m.Payload)
	}
	h.do("DELETE", "/api/tasks/"+task.ID, a.Token, nil, nil)
	m = next(t, c)
	if m.ID != "3" || !strings.Contains(string(m.Payload), task.ID) {
		t.Fatalf("taskDeleted: %+v %s", m, m.Payload)
	}

	// metrics report subscribers
	var met Metrics
	h.do("GET", "/metrics", "", nil, &met)
	if met.Subscribers < 3 || met.Requests == 0 || met.DBQueries != 1 {
		t.Fatalf("metrics %+v", met)
	}
}

func TestSubscriptionAuth(t *testing.T) {
	h := newHarness(t, "production")

	// anonymous connection is accepted, but subscribing requires auth
	c, err := dialWS(t, h, "", "http://allowed.test")
	if err != nil {
		t.Fatal(err)
	}
	for i, q := range []string{`subscription{ taskCreated{ id } }`, `subscription{ taskUpdated{ id } }`, `subscription{ taskDeleted }`} {
		id := string(rune('a' + i))
		subscribe(t, c, id, q)
		m := next(t, c)
		// gqlgen reports resolver errors as an "error" or a "next" carrying errors
		if (m.Type != "error" && m.Type != "next") || !strings.Contains(string(m.Payload), "UNAUTHENTICATED") {
			t.Fatalf("%s: %+v %s", q, m, m.Payload)
		}
	}
	c.Close()

	// invalid token is rejected at connection_init
	if _, err := dialWS(t, h, "garbage", "http://allowed.test"); err == nil {
		t.Fatal("invalid token accepted")
	}
	// unknown origin is rejected in production
	if _, err := dialWS(t, h, "", "http://evil.test"); err == nil {
		t.Fatal("foreign origin accepted")
	}
}

func TestSubscriptionContextCancel(t *testing.T) {
	h := newHarness(t, "development")
	a := h.register("cancel@x.com")
	c, err := dialWS(t, h, a.Token, "")
	if err != nil {
		t.Fatal(err)
	}
	subscribe(t, c, "1", `subscription{ taskUpdated{ id } }`)
	subscribe(t, c, "2", `subscription{ taskDeleted }`)
	time.Sleep(100 * time.Millisecond)
	// publish events nobody reads, then hang up: the server must clean up
	for i := 0; i < 3; i++ {
		var task taskJSON
		h.do("POST", "/api/tasks", a.Token, map[string]string{"title": "x"}, &task)
		h.do("PATCH", "/api/tasks/"+task.ID, a.Token, map[string]string{"status": "DONE"}, nil)
		h.do("DELETE", "/api/tasks/"+task.ID, a.Token, nil, nil)
	}
	c.Close()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		var met Metrics
		h.do("GET", "/metrics", "", nil, &met)
		if met.Subscribers == 0 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("subscribers not cleaned up after disconnect")
}

// ---------- operational endpoints ----------

func TestOperationalEndpoints(t *testing.T) {
	h := newHarness(t, "development")
	res, err := http.Get(h.srv.URL + "/health")
	if err != nil || res.StatusCode != 200 {
		t.Fatal("health")
	}
	if res.Header.Get("X-Content-Type-Options") != "nosniff" || res.Header.Get("X-Frame-Options") != "DENY" {
		t.Fatal("security headers missing")
	}
	if code := h.do("GET", "/ready", "", nil, nil); code != 200 {
		t.Fatalf("ready %d", code)
	}
	h.ready = errors.New("db down")
	var body map[string]string
	if code := h.do("GET", "/ready", "", nil, &body); code != 503 || body["error"] != "db down" {
		t.Fatalf("not ready %d %v", code, body)
	}
	res, _ = http.Get(h.srv.URL + "/")
	if res.StatusCode != 200 {
		t.Fatal("playground should be served in development")
	}

	// CORS preflight
	req, _ := http.NewRequest("OPTIONS", h.srv.URL+"/query", nil)
	req.Header.Set("Origin", "http://allowed.test")
	req.Header.Set("Access-Control-Request-Method", "POST")
	res, _ = http.DefaultClient.Do(req)
	if res.Header.Get("Access-Control-Allow-Origin") != "http://allowed.test" {
		t.Fatal("CORS preflight not answered")
	}

	prod := newHarness(t, "production")
	res, _ = http.Get(prod.srv.URL + "/")
	if res.StatusCode != 404 {
		t.Fatalf("playground should be disabled in production, got %d", res.StatusCode)
	}
}

func TestMetricsWithoutOptionalDeps(t *testing.T) {
	jwt := auth.NewJWTManager("s", "r")
	users := memstore.NewUsers()
	srv := httptest.NewServer(New(Deps{
		Env: "production", JWT: jwt,
		AuthSvc: service.NewAuthService(users, jwt),
		UserSvc: service.NewUserService(users, nil, time.Minute),
		TaskSvc: service.NewTaskService(memstore.NewTasks(), users, nil, nil, time.Minute),
		Broker:  pubsub.NewBroker(),
	}))
	defer srv.Close()
	res, err := http.Get(srv.URL + "/metrics")
	if err != nil || res.StatusCode != 200 {
		t.Fatal("metrics")
	}
	res, _ = http.Get(srv.URL + "/ready")
	if res.StatusCode != 200 {
		t.Fatal("ready without checker should be 200")
	}
	if !originAllowed(Deps{AllowedOrigins: []string{"*"}}, "http://any") {
		t.Fatal("wildcard origin")
	}
}
