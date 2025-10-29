package main

import (
	"testing"
	"taskboard/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to test database")
	}

	// Migrate tables
	db.AutoMigrate(&models.Task{}, &models.User{})
	return db
}

func TestTaskRepository(t *testing.T) {
	db := setupTestDB()
	repo := NewTaskRepository(db)

	t.Run("Create Task", func(t *testing.T) {
		task := &models.Task{
			Title:       "Test Task",
			Description: "Test Description",
			Status:      "todo",
			AssigneeID:  1,
		}

		err := repo.Create(task)
		assert.NoError(t, err)
		assert.NotZero(t, task.ID)
	})

	t.Run("Get Task By ID", func(t *testing.T) {
		// First create a task
		task := &models.Task{
			Title: "Get Test Task",
			Status: "todo",
		}
		repo.Create(task)

		// Then retrieve it
		foundTask, err := repo.GetByID(task.ID)
		assert.NoError(t, err)
		assert.Equal(t, task.Title, foundTask.Title)
	})

	t.Run("Update Task", func(t *testing.T) {
		task := &models.Task{
			Title: "Update Test Task",
			Status: "todo",
		}
		repo.Create(task)

		task.Status = "in-progress"
		err := repo.Update(task)
		assert.NoError(t, err)

		updatedTask, _ := repo.GetByID(task.ID)
		assert.Equal(t, "in-progress", updatedTask.Status)
	})

	t.Run("Delete Task", func(t *testing.T) {
		task := &models.Task{
			Title: "Delete Test Task",
			Status: "todo",
		}
		repo.Create(task)

		err := repo.Delete(task.ID)
		assert.NoError(t, err)

		_, err = repo.GetByID(task.ID)
		assert.Error(t, err)
	})

	t.Run("Get All Tasks", func(t *testing.T) {
		// Clear existing tasks
		db.Where("1=1").Delete(&models.Task{})

		// Create test tasks
		tasks := []models.Task{
			{Title: "Task 1", Status: "todo"},
			{Title: "Task 2", Status: "in-progress"},
		}

		for i := range tasks {
			repo.Create(&tasks[i])
		}

		allTasks, err := repo.GetAll()
		assert.NoError(t, err)
		assert.Len(t, allTasks, 2)
	})
}

func TestUserRepository(t *testing.T) {
	db := setupTestDB()
	userRepo := NewUserRepository(db)

	t.Run("Create User", func(t *testing.T) {
		user := &models.User{
			Name:  "Test User",
			Email: "test@example.com",
		}

		err := userRepo.Create(user)
		assert.NoError(t, err)
		assert.NotZero(t, user.ID)
	})

	t.Run("Get User By ID", func(t *testing.T) {
		user := &models.User{
			Name:  "Get Test User",
			Email: "gettest@example.com",
		}
		userRepo.Create(user)

		foundUser, err := userRepo.GetByID(user.ID)
		assert.NoError(t, err)
		assert.Equal(t, user.Name, foundUser.Name)
	})

	t.Run("Get All Users", func(t *testing.T) {
		users, err := userRepo.GetAll()
		assert.NoError(t, err)
		assert.NotNil(t, users)
	})
}

// Mock repositories for testing
type TaskRepository interface {
	Create(task *models.Task) error
	GetByID(id uint) (*models.Task, error)
	Update(task *models.Task) error
	Delete(id uint) error
	GetAll() ([]models.Task, error)
}

type UserRepository interface {
	Create(user *models.User) error
	GetByID(id uint) (*models.User, error)
	GetAll() ([]models.User, error)
}