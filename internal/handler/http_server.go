package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gulovv/calculator_with_auth/internal/compute"
	"github.com/gulovv/calculator_with_auth/internal/dag"
	"github.com/gulovv/calculator_with_auth/pkg/model"
	"golang.org/x/crypto/bcrypt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)
func CalculateHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        login, err := validateToken(r)  // Проверка токена и извлечение логина
        if err != nil {
            http.Error(w, err.Error(), http.StatusUnauthorized)
            return
        }

        var req model.Request
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid JSON", http.StatusBadRequest)
            return
        }

        if strings.TrimSpace(req.Expression) == "" {
            http.Error(w, "Empty expression", http.StatusUnprocessableEntity)
            return
        }

        fmt.Println("Новый запрос на вычисление:", req.Expression)

        // Получаем user_id из базы данных по логину
        var userID int
        err = db.QueryRow("SELECT id FROM users WHERE login = $1", login).Scan(&userID)
        if err != nil {
            if err == sql.ErrNoRows {
                http.Error(w, "User not found", http.StatusUnauthorized)
            } else {
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
            return
        }

        // Токенизация и обработка выражения
        tokens, err := dag.Tokenize(req.Expression)
        if err != nil {
            http.Error(w, "Invalid expression tokens", http.StatusUnprocessableEntity)
            return
        }

        rpn, err := dag.ToRPN(tokens)
        if err != nil {
            http.Error(w, "Failed to convert to RPN", http.StatusUnprocessableEntity)
            return
        }

        known, dagNodes, finalID, err := dag.BuildDAG(rpn)
        if err != nil {
            http.Error(w, "Invalid DAG structure", http.StatusUnprocessableEntity)
            return
        }

        // Вычисление результата
        result := compute.Orchestrate(dagNodes, known, finalID, 2)

        // Генерация уникального ID для вычисления
        id := uuid.New().String()

        // Добавление выражения в базу данных с привязкой user_id
        _, err = db.Exec(`INSERT INTO calculations (id, expression, result, user_id) VALUES ($1, $2, $3, $4)`,
            id, req.Expression, result, userID)
        if err != nil {
            fmt.Println("Ошибка при сохранении в БД:", err)
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }

        // Ответ с ID вычисления
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(map[string]string{"id": id})
    }
}


func GetExpressionsHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        login, err := validateToken(r)  // Проверка токена
        if err != nil {
            http.Error(w, err.Error(), http.StatusUnauthorized)
            return
        }

        // Получаем user_id из базы данных по логину
        var userID int
        err = db.QueryRow("SELECT id FROM users WHERE login = $1", login).Scan(&userID)
        if err != nil {
            if err == sql.ErrNoRows {
                http.Error(w, "User not found", http.StatusUnauthorized)
            } else {
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
            return
        }

        rows, err := db.Query(`SELECT id, expression, result FROM calculations WHERE user_id = $1`, userID)
        if err != nil {
            log.Println("Ошибка при запросе к БД:", err)
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }
        defer rows.Close()

        var expressions []model.ExpressionResponse

        for rows.Next() {
            var expr model.ExpressionResponse
            if err := rows.Scan(&expr.ID, &expr.Expression, &expr.Result); err != nil {
                log.Println("Ошибка при чтении строки:", err)
                continue
            }
            expressions = append(expressions, expr)
        }

        response := model.ExpressionsListResponse{Expressions: expressions}

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(response)
    }
}


func GetExpressionByIDHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        login, err := validateToken(r)  // Проверка токена
        if err != nil {
            http.Error(w, err.Error(), http.StatusUnauthorized)
            return
        }

        // Получаем user_id из базы данных по логину
        var userID int
        err = db.QueryRow("SELECT id FROM users WHERE login = $1", login).Scan(&userID)
        if err != nil {
            if err == sql.ErrNoRows {
                http.Error(w, "User not found", http.StatusUnauthorized)
            } else {
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
            return
        }

        parts := strings.Split(r.URL.Path, "/")
        if len(parts) < 5 {
            http.Error(w, "Invalid URL", http.StatusBadRequest)
            return
        }
        id := parts[4]

        var expr model.ExpressionResponse
        err = db.QueryRow("SELECT id, expression, result FROM calculations WHERE id = $1 AND user_id = $2", id, userID).Scan(
            &expr.ID, &expr.Expression, &expr.Result,
        )
        if err == sql.ErrNoRows {
            http.Error(w, "Expression not found", http.StatusNotFound)
            return
        } else if err != nil {
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]model.ExpressionResponse{
            "expression": expr,
        })
    }
}



func validateToken(r *http.Request) (string, error) {
    tokenString := r.Header.Get("Authorization")
    if tokenString == "" {
        return "", errors.New("authorization header is missing")
    }

    parts := strings.Split(tokenString, " ")
    if len(parts) != 2 || parts[0] != "Bearer" {
        return "", errors.New("invalid token format")
    }

    token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("invalid signing method")
        }
        return jwtKey, nil
    })
    if err != nil {
        return "", err
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok || !token.Valid {
        return "", errors.New("invalid token")
    }

    login, ok := claims["login"].(string)
    if !ok {
        return "", errors.New("login not found in token")
    }

    return login, nil
}


func RegisterHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req model.RegisterRequest

        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "неверный формат запроса", http.StatusBadRequest)
            return
        }

        if req.Login == "" || req.Password == "" {
            http.Error(w, "логин и пароль обязательны", http.StatusBadRequest)
            return
        }

        // Проверка на наличие уже зарегистрированного пользователя
        var existingUser string
        err := db.QueryRow(`SELECT login FROM users WHERE login = $1`, req.Login).Scan(&existingUser)
        if err == nil {
            http.Error(w, "пользователь с таким логином уже существует", http.StatusBadRequest)
            return
        } else if err != sql.ErrNoRows {
            http.Error(w, "ошибка базы данных", http.StatusInternalServerError)
            return
        }

        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
        if err != nil {
            http.Error(w, "ошибка хеширования пароля", http.StatusInternalServerError)
            return
        }

        var userID int
		err = db.QueryRow(`INSERT INTO users (login, hashed_password) VALUES ($1, $2) RETURNING id`, req.Login, string(hashedPassword)).Scan(&userID)
		if err != nil {
			http.Error(w, "ошибка при сохранении пользователя: "+err.Error(), http.StatusInternalServerError)
			return
		}

        response := map[string]string{
            "message": "пользователь успешно зарегистрирован",
        }
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(response)
    }
}

func LoginHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req model.LoginRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "неверный формат запроса", http.StatusBadRequest)
            return
        }

        if req.Login == "" || req.Password == "" {
            http.Error(w, "логин и пароль обязательны", http.StatusBadRequest)
            return
        }

        var hashedPassword string
        err := db.QueryRow(`SELECT hashed_password FROM users WHERE login = $1`, req.Login).Scan(&hashedPassword)
        if err == sql.ErrNoRows {
            http.Error(w, "пользователь не найден", http.StatusUnauthorized)
            return
        } else if err != nil {
            http.Error(w, "ошибка базы данных", http.StatusInternalServerError)
            return
        }

        err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password))
        if err != nil {
            http.Error(w, "неверный пароль", http.StatusUnauthorized)
            return
        }

        tokenString, err := generateJWTToken(req.Login)
        if err != nil {
            http.Error(w, "ошибка создания токена", http.StatusInternalServerError)
            return
        }

        response := map[string]string{
            "token": tokenString,
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
    }
}