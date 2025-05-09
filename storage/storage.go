package storage

import (
    "database/sql"
    "fmt"
    "log"
    "os"
	"time"

    _ "github.com/lib/pq"
)
func InitDB() (*sql.DB, error) {
    dbUser := os.Getenv("DB_USER")
    dbPassword := os.Getenv("DB_PASSWORD")
    dbName := os.Getenv("DB_NAME")
    dbHost := os.Getenv("DB_HOST")
    dbPort := os.Getenv("DB_PORT")

    connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable", dbUser, dbPassword, dbHost, dbPort)

    var systemDB *sql.DB
    var err error

    // Повторные попытки подключения к системной БД
    for i := 0; i < 10; i++ {
        systemDB, err = sql.Open("postgres", connStr)
        if err == nil {
            err = systemDB.Ping()
            if err == nil {
                break
            }
        }
        log.Printf("Попытка %d: Ожидание подключения к БД... (%v)", i+1, err)
        time.Sleep(3 * time.Second)
    }

    if err != nil {
        return nil, fmt.Errorf("не удалось подключиться к системной БД: %w", err)
    }
    defer systemDB.Close()

    var exists bool
    query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '%s');", dbName)
    err = systemDB.QueryRow(query).Scan(&exists)
    if err != nil {
        return nil, fmt.Errorf("ошибка при проверке существования базы данных: %w", err)
    }

    if !exists {
        _, err = systemDB.Exec(fmt.Sprintf("CREATE DATABASE %s;", dbName))
        if err != nil {
            return nil, fmt.Errorf("ошибка при создании базы данных: %w", err)
        }
        log.Println("База данных создана")
    } else {
        log.Println("База данных уже существует, продолжаем")
    }

    // Подключение к целевой БД
    dbConnStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)
    db, err := sql.Open("postgres", dbConnStr)
    if err != nil {
        return nil, fmt.Errorf("ошибка подключения к базе данных %s: %w", dbName, err)
    }

    // Создание таблиц
    userQuery := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        login TEXT UNIQUE NOT NULL,
        hashed_password TEXT NOT NULL
    );`
    _, err = db.Exec(userQuery)
    if err != nil {
        return nil, fmt.Errorf("ошибка при создании таблицы users: %w", err)
    }
    log.Println("Таблица users проверена/создана")

    calcQuery := `
    CREATE TABLE IF NOT EXISTS calculations (
        id TEXT PRIMARY KEY, 
        expression TEXT NOT NULL,
        result REAL NOT NULL,
        user_id INTEGER REFERENCES users(id)
    );`
    _, err = db.Exec(calcQuery)
    if err != nil {
        return nil, fmt.Errorf("ошибка при создании таблицы calculations: %w", err)
    }
    log.Println("Таблица calculations проверена/создана")

    return db, nil
}