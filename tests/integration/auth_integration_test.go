package integration

import (
	"bytes"
	"encoding/json"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"url/short/configs"
	"url/short/internal/auth"
	"url/short/internal/user"
	db "url/short/pkg/db"
)

// TestDB представляет тестовую базу данных
type TestDB struct {
	*gorm.DB
}

// SetupTestDB создает тестовую БД
func SetupTestDB(t *testing.T) *TestDB {
	// Используем тестовую БД с предоставленными учетными данными
	dsn := "host=localhost user=user password=password dbname=link_test port=5432 sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("Skipping integration test: failed to connect to database: %v", err)
	}

	return &TestDB{db}
}

// Cleanup очищает тестовые данные
func (tdb *TestDB) Cleanup() {
	tdb.Unscoped().Where("email LIKE ?", "test_%").Delete(&user.User{})
}

// CreateTestUser создает тестового пользователя
func (tdb *TestDB) CreateTestUser(email, password, name string) *user.User {
	user := &user.User{
		Email:    email,
		Password: password,
		Name:     name,
	}
	tdb.Create(user)
	return user
}

func TestAuthIntegration_LoginSuccess(t *testing.T) {
	// Настройка тестовой БД
	testDB := SetupTestDB(t)
	defer testDB.Cleanup()

	// Создаем тестового пользователя
	testUser := testDB.CreateTestUser(
		"test_login@example.com",
		"$2a$10$xwLLgG77tJ5x9hWAXJrk0OFq/bpY4i9pojqsmxLyznn45A5.COVb6", // хеш для "123"
		"Test User",
	)

	// Создаем тестовый сервер
	app := createTestApp(testDB.DB)
	ts := httptest.NewServer(app)
	defer ts.Close()

	// Подготавливаем запрос
	loginReq := auth.LoginRequest{
		Email:    testUser.Email,
		Password: "123",
	}
	data, err := json.Marshal(loginReq)
	if err != nil {
		t.Fatal(err)
	}

	// Выполняем запрос
	res, err := http.Post(ts.URL+"/auth/login", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	// Проверяем результат
	if res.StatusCode != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, res.StatusCode)
	}

	// Проверяем ответ
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	var loginResp auth.LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		t.Fatal(err)
	}

	if loginResp.Token == "" {
		t.Fatal("Expected token in response")
	}
}

func TestAuthIntegration_LoginFailed(t *testing.T) {
	// Настройка тестовой БД
	testDB := SetupTestDB(t)
	defer testDB.Cleanup()

	// Создаем тестового пользователя
	testDB.CreateTestUser(
		"test_failed@example.com",
		"$2a$10$xwLLgG77tJ5x9hWAXJrk0OFq/bpY4i9pojqsmxLyznn45A5.COVb6",
		"Test User",
	)

	// Создаем тестовый сервер
	app := createTestApp(testDB.DB)
	ts := httptest.NewServer(app)
	defer ts.Close()

	// Подготавливаем запрос с неверными данными
	loginReq := auth.LoginRequest{
		Email:    "test_failed@example.com",
		Password: "wrong_password",
	}
	data, err := json.Marshal(loginReq)
	if err != nil {
		t.Fatal(err)
	}

	// Выполняем запрос
	res, err := http.Post(ts.URL+"/auth/login", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	// Проверяем результат
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("Expected status %d, got %d", http.StatusUnauthorized, res.StatusCode)
	}
}

func TestAuthIntegration_RegisterSuccess(t *testing.T) {
	// Настройка тестовой БД
	testDB := SetupTestDB(t)
	defer testDB.Cleanup()

	// Создаем тестовый сервер
	app := createTestApp(testDB.DB)
	ts := httptest.NewServer(app)
	defer ts.Close()

	// Подготавливаем запрос
	registerReq := auth.RegisterRequest{
		Email:    "test_register@example.com",
		Password: "password123",
		Name:     "Test User",
	}
	data, err := json.Marshal(registerReq)
	if err != nil {
		t.Fatal(err)
	}

	// Выполняем запрос
	res, err := http.Post(ts.URL+"/auth/register", "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	// Проверяем результат
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status %d, got %d", http.StatusCreated, res.StatusCode)
	}

	// Проверяем, что пользователь создан в БД
	var user user.User
	result := testDB.Where("email = ?", registerReq.Email).First(&user)
	if result.Error != nil {
		t.Fatalf("User should be created in database: %v", result.Error)
	}
}

// createTestApp создает тестовое приложение с переданной БД
func createTestApp(gormDB *gorm.DB) http.Handler {
	// Создаем тестовую конфигурацию
	conf := &configs.Config{
		Db: configs.Dbconfig{
			Dsn: "test_dsn",
		},
		Auth: configs.Authconfig{
			Secret: "test_secret_key_at_least_32_characters_long",
		},
	}

	// Создаем обертку базы данных
	dbWrapper := &db.DB{DB: gormDB}

	// Инициализируем репозитории
	userRepo := user.NewUserRepository(dbWrapper)

	// Инициализируем сервисы
	authService := auth.NewAuthService(userRepo)

	// Создаем маршрутизатор
	r := http.NewServeMux()

	// Регистрируем хендлеры
	auth.NewAuthHandler(r, auth.AuthHandlerDeps{
		Config:      conf,
		AuthService: authService,
	})

	return r
}
