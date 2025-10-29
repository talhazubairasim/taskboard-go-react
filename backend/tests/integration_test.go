package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"taskboard/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestTaskIntegration(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := setupRouter()
	
	// Test Create Task
	task := models.Task{
		Title:       "Integration Test Task",
		Description: "This is a test task for integration testing",
		Status:      "todo",
		AssigneeID:  1,
	}

	jsonData, _ := json.Marshal(task)
	req, _ := http.NewRequest("POST", "/api/tasks", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var createdTask models.Task
	json.Unmarshal(w.Body.Bytes(), &createdTask)

	assert.Equal(t, task.Title, createdTask.Title)
	assert.Equal(t, task.Description, createdTask.Description)

	// Test Get Task
	req, _ = http.NewRequest("GET", "/api/tasks/"+string(createdTask.ID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test Update Task
	updatedTask := createdTask
	updatedTask.Status = "in-progress"

	jsonData, _ = json.Marshal(updatedTask)
	req, _ = http.NewRequest("PUT", "/api/tasks/"+string(createdTask.ID), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Test Delete Task
	req, _ = http.NewRequest("DELETE", "/api/tasks/"+string(createdTask.ID), nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUserIntegration(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	// Test Get Users
	req, _ := http.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var users []models.User
	json.Unmarshal(w.Body.Bytes(), &users)
	assert.NotNil(t, users)
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	
	// Add your routes here
	// router.POST("/api/tasks", CreateTask)
	// router.GET("/api/tasks/:id", GetTask)
	// router.PUT("/api/tasks/:id", UpdateTask)
	// router.DELETE("/api/tasks/:id", DeleteTask)
	// router.GET("/api/users", GetUsers)
	
	return router
}