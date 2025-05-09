package test

import (
    "github.com/gulovv/calculator_with_auth/internal/handler"
    "github.com/gulovv/calculator_with_auth/pkg/model"
    "database/sql"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    _ "github.com/lib/pq"
    "github.com/google/uuid"
)

func setupTestDB(t *testing.T) *sql.DB {
    db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/calculations?sslmode=disable")
    if err != nil {
        t.Fatalf("Ошибка подключения к БД: %v", err)
    }

    _, err = db.Exec("DELETE FROM calculations") // Очистка БД перед каждым тестом
    if err != nil {
        t.Fatalf("Ошибка очистки таблицы: %v", err)
    }

    return db
}

func TestGetExpressionsHandler(t *testing.T) {
    db := setupTestDB(t)

    // Добавляем одно выражение для теста
    id := uuid.New().String()
    _, err := db.Exec(`INSERT INTO calculations (id, expression, result) VALUES ($1, $2, $3)`, id, "3 + 5 * 2", 13.0)
    if err != nil {
        t.Fatalf("Ошибка при добавлении данных в БД: %v", err)
    }

    req := httptest.NewRequest(http.MethodGet, "/api/v1/expressions", nil)
    w := httptest.NewRecorder()

    handler.GetExpressionsHandler(db)(w, req)

    res := w.Result()
    defer res.Body.Close()

    if res.StatusCode != http.StatusOK {
        t.Fatalf("Ожидался статус 200, но получен %d", res.StatusCode)
    }

    var response model.ExpressionsListResponse
    if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
        t.Fatalf("Ошибка декодирования ответа: %v", err)
    }

    if len(response.Expressions) != 1 {
        t.Errorf("Ожидалось 1 выражение, получено %d", len(response.Expressions))
    }

    if response.Expressions[0].Expression != "3 + 5 * 2" {
        t.Errorf("Ожидалось выражение '3 + 5 * 2', получено '%s'", response.Expressions[0].Expression)
    }

    if response.Expressions[0].Result != 13 {
        t.Errorf("Ожидался результат 13, получен %f", response.Expressions[0].Result)
    }
}

func TestGetExpressionByIDHandler(t *testing.T) {
    db := setupTestDB(t)

    id := uuid.New().String()
    _, err := db.Exec(`INSERT INTO calculations (id, expression, result) VALUES ($1, $2, $3)`, id, "3 + 5 * 2", 13.0)
    if err != nil {
        t.Fatalf("Ошибка при добавлении данных в БД: %v", err)
    }

    req := httptest.NewRequest(http.MethodGet, "/api/v1/expressions/"+id, nil)
    w := httptest.NewRecorder()

    handler.GetExpressionByIDHandler(db)(w, req)

    res := w.Result()
    defer res.Body.Close()

    if res.StatusCode != http.StatusOK {
        t.Fatalf("Ожидался статус 200, но получен %d", res.StatusCode)
    }

    var body map[string]model.ExpressionResponse
    if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
        t.Fatalf("Ошибка декодирования ответа: %v", err)
    }

    expr := body["expression"]
    if expr.ID != id || expr.Expression != "3 + 5 * 2" || expr.Result != 13 {
        t.Errorf("Неверные данные: %+v", expr)
    }
}