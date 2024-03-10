package service

import (
	"context"
	"customer-service/db"
	_ "errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testSqlDB *sqlx.DB

func TestMain(m *testing.M) {
	container, testDB, err := createContainer("postgres")
	if err != nil {
		log.Fatal(err)
	}
	defer testDB.Close()
	defer container.Terminate(context.Background())

	testSqlDB = testDB

	content, err := os.ReadFile("setup.sql")
	if err != nil {
		log.Fatal(err)
	}

	_, err = testDB.Exec(string(content))
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}

func TestCreateCustomer(t *testing.T) {
	// Given
	mockDB := &db.PostgresDB{
		DB: testSqlDB,
	}

	// When
	router := gin.New()
	app := GetApp(mockDB)
	router.POST("/customers", app.PostHandler)

	// Then case 1: Valid input
	resp := performRequest(router, "POST", "/customers", `{"name": "John Doe", "email": "john.doe@example.com", "address": "123 Main St"}`)
	assert.Equal(t, http.StatusCreated, resp.Code)

	// Then case 2: Missing email
	resp = performRequest(router, "POST", "/customers", `{"name": "John Doe", "address": "123 Main St"}`)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestGetCustomer(t *testing.T) {
	// Given
	mockDB := &db.PostgresDB{
		DB: testSqlDB,
	}

	// When
	router := gin.New()
	app := GetApp(mockDB)
	router.GET("/customers/:customerId", app.GetHandler)

	// Then case 1: Valid input
	resp := performRequest(router, "GET", "/customers/1", "")
	assert.Equal(t, http.StatusOK, resp.Code)

	// Then case 2: Invalid customer ID
	resp = performRequest(router, "GET", "/customers/invalid", "")
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	// Then case 3: Customer not found
	resp = performRequest(router, "GET", "/customers/2", "")
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestUpdateCustomer(t *testing.T) {
	// Given
	mockDB := &db.PostgresDB{
		DB: testSqlDB,
	}

	// When
	router := gin.New()
	app := GetApp(mockDB)
	router.PUT("/customers/:customerId", app.PutHandler)

	// Then case 1: Valid input
	resp := performRequest(router, "PUT", "/customers/1", `{"name": "Updated Name", "address": "Updated Address"}`)
	assert.Equal(t, http.StatusOK, resp.Code)

	// Then case 2: Invalid customer ID
	resp = performRequest(router, "PUT", "/customers/invalid", "")
	assert.Equal(t, http.StatusBadRequest, resp.Code)

	// Then case 3: No modification (empty request body)
	resp = performRequest(router, "PUT", "/customers/1", `{}`)
	assert.Equal(t, http.StatusNotModified, resp.Code)
}

func TestDeleteCustomer(t *testing.T) {
	// Given
	mockDB := &db.PostgresDB{
		DB: testSqlDB,
	}

	// When
	router := gin.New()
	app := GetApp(mockDB)
	router.DELETE("/customers/:customerId", app.DeleteHandler)

	// Then case 1: Valid input
	resp := performRequest(router, "DELETE", "/customers/1", "")
	assert.Equal(t, http.StatusNoContent, resp.Code)

	// Then case 2: Invalid customer ID
	resp = performRequest(router, "DELETE", "/customers/invalid", "")
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func performRequest(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	return resp
}

func createContainer(databaseName string) (testcontainers.Container, *sqlx.DB, error) {
	port := "5432/tcp"

	var env = map[string]string{
		"POSTGRES_USER":     databaseName,
		"POSTGRES_PASSWORD": "password",
		"POSTGRES_DB":       databaseName,
	}

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "postgres:12.5",
			ExposedPorts: []string{port},
			Cmd:          []string{"postgres", "-c", "fsync=off"},
			Env:          env,
			Name:         databaseName + uuid.New().String(),
			WaitingFor: wait.ForSQL(nat.Port(port), "pgx", func(host string, port nat.Port) string {
				return fmt.Sprintf("postgres://%s:password@%s:%s/%s?sslmode=disable", databaseName, host, port.Port(), databaseName)
			}),
		},
		Started: true,
	}

	container, err := testcontainers.GenericContainer(context.Background(), req)
	if err != nil {
		log.Fatal(err)
	}

	mappedPort, err := container.MappedPort(context.Background(), nat.Port(port))
	if err != nil {
		log.Fatal(err)
		return nil, nil, err
	}

	url := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s statement_cache_mode=describe", "localhost", mappedPort.Port(), databaseName, "password", databaseName)
	db, err := sqlx.Open("pgx", url)
	if err != nil {
		fmt.Fprintln(os.Stdout, "stuff here")
		return nil, nil, err
	}

	return container, db, nil
}
