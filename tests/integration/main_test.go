package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"ssubench/internal/app"
	"ssubench/internal/config"
	"ssubench/internal/domain"
	"strings"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/bcrypt"
)

var (
	testPool *pgxpool.Pool
	router   http.Handler
	cfg      *config.Config
)

func TestMain(m *testing.M) {
	code := runTestMain(m)

	os.Exit(code)
}

func runTestMain(m *testing.M) int {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dbName := "ssubench_test"
	dbUser := "user"
	dbPassword := "password"

	postgresContainer, err := postgres.Run(ctx,
		"postgres:18.3-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		log.Printf("Failed to start container: %v", err)
		return 1
	}

	defer postgresContainer.Terminate(ctx)

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Print(err)
		return 1
	}

	testPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		log.Print(err)
		return 1
	}

	defer testPool.Close()

	err = runMigrations(ctx, connStr)
	if err != nil {
		log.Print(err)
		return 1
	}

	cfg, err = config.Load()
	if err != nil {
		log.Print(err)
		return 1
	}

	router = app.SetupHandler(testPool, cfg, false)

	return m.Run()
}

func runMigrations(ctx context.Context, connStr string) error {
	idx := strings.IndexRune(connStr, ':')
	driverURL := fmt.Sprintf("pgx5%s", connStr[idx:])

	mig, err := migrate.New("file://../../migrations", driverURL)
	if err != nil {
		return err
	}

	err = mig.Up()

	if err != nil {
		return err
	}

	return nil
}

func clearDB(t *testing.T) error {
	t.Helper()

	_, err := testPool.Exec(context.Background(), "TRUNCATE users, tasks, bids, payments RESTART IDENTITY CASCADE")
	if err != nil {
		return err
	}

	return nil
}

func registerAndLogin(t *testing.T, username, password, role string) (string, error) {
	t.Helper()

	registerBody, err := json.Marshal(map[string]string{
		"username": username,
		"password": password,
		"role":     role,
	})

	if err != nil {
		return "", err
	}

	request := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(registerBody))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	if w.Code != http.StatusCreated {
		return "", fmt.Errorf("Registration failed")
	}

	loginBody, err := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})

	if err != nil {
		return "", err
	}

	request = httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(loginBody))
	w = httptest.NewRecorder()

	router.ServeHTTP(w, request)

	if w.Code != http.StatusOK {
		return "", fmt.Errorf("Login failed")
	}

	var response struct {
		Token string `json:"token"`
	}

	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		return "", err
	}

	return response.Token, nil
}

func doRequest(method, path string, body any, token string) (*httptest.ResponseRecorder, error) {
	var buffer bytes.Buffer
	if body != nil {
		err := json.NewEncoder(&buffer).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	request := httptest.NewRequest(method, path, &buffer)
	request.Header.Set("Content-Type", "application/json")

	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, request)
	return w, nil
}

func addAdminAndLogin(t *testing.T, username, password string) (string, error) {
	t.Helper()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cfg.PasswordCost)

	if err != nil {
		return "", err
	}

	user := &domain.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Role:         domain.RoleAdmin,
		Status:       domain.StatusActive,
		Balance:      cfg.StartingBalance,
	}

	query := `INSERT INTO users (username, password_hash, role, status, balance)
			  VALUES ($1, $2, $3, $4, $5)`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = testPool.Exec(ctx, query,
		user.Username,
		user.PasswordHash,
		user.Role,
		user.Status,
		user.Balance,
	)

	if err != nil {
		return "", err
	}

	loginBody, err := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})

	if err != nil {
		return "", err
	}

	request := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(loginBody))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	if w.Code != http.StatusOK {
		return "", fmt.Errorf("Login failed")
	}

	var response struct {
		Token string `json:"token"`
	}

	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		return "", err
	}

	return response.Token, nil
}
