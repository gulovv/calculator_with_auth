package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"github.com/gulovv/calculator_with_auth/internal/compute"
	"github.com/gulovv/calculator_with_auth/internal/dag"
	"github.com/gulovv/calculator_with_auth/internal/handler/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type CalculatorServer struct {
    pb.UnimplementedCalculatorServiceServer
    DB *sql.DB
}

// Реализация метода Calculate
func (s *CalculatorServer) Calculate(ctx context.Context, req *pb.CalculationRequest) (*pb.CalculationResponse, error) {
    login, err := validateTokenFromContext(ctx)  // Заменим validateToken на функцию для gRPC контекста
    if err != nil {
        return nil, status.Error(codes.Unauthenticated, err.Error())
    }

    // Получаем user_id из базы данных по логину
    var userID int
    err = s.DB.QueryRow("SELECT id FROM users WHERE login = $1", login).Scan(&userID)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, status.Error(codes.Unauthenticated, "User not found")
        }
        return nil, status.Error(codes.Internal, "Internal Server Error")
    }

    // Токенизация и обработка выражения
    tokens, err := dag.Tokenize(req.Expression)
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, "Invalid expression tokens")
    }

    rpn, err := dag.ToRPN(tokens)
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, "Failed to convert to RPN")
    }

    known, dagNodes, finalID, err := dag.BuildDAG(rpn)
    if err != nil {
        return nil, status.Error(codes.InvalidArgument, "Invalid DAG structure")
    }

    // Вычисление результата
    result := compute.Orchestrate(dagNodes, known, finalID, 2)

    // Генерация уникального ID для вычисления
    id := uuid.New().String()

    // Добавление выражения в базу данных с привязкой user_id
    _, err = s.DB.Exec(`INSERT INTO calculations (id, expression, result, user_id) VALUES ($1, $2, $3, $4)`,
        id, req.Expression, result, userID)
    if err != nil {
        log.Println("Ошибка при сохранении в БД:", err)
        return nil, status.Error(codes.Internal, "Internal Server Error")
    }

    return &pb.CalculationResponse{
		Id:     id,
		Result: fmt.Sprintf("%f", result),
	}, nil
}
	


// Реализация метода GetExpressions
func (s *CalculatorServer) GetExpressions(ctx context.Context, req *pb.GetExpressionsRequest) (*pb.GetExpressionsResponse, error) {
    login, err := validateTokenFromContext(ctx)
    if err != nil {
        return nil, status.Error(codes.Unauthenticated, err.Error())
    }

    // Получаем user_id из базы данных по логину
    var userID int
    err = s.DB.QueryRow("SELECT id FROM users WHERE login = $1", login).Scan(&userID)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, status.Error(codes.Unauthenticated, "User not found")
        }
        return nil, status.Error(codes.Internal, "Internal Server Error")
    }

    rows, err := s.DB.Query(`SELECT id, expression, result FROM calculations WHERE user_id = $1`, userID)
    if err != nil {
        return nil, status.Error(codes.Internal, "Internal Server Error")
    }
    defer rows.Close()

    var expressions []*pb.Expression
    for rows.Next() {
        var expr pb.Expression
        if err := rows.Scan(&expr.Id, &expr.Expression, &expr.Result); err != nil {
            log.Println("Ошибка при чтении строки:", err)
            continue
        }
        expressions = append(expressions, &expr)
    }

    return &pb.GetExpressionsResponse{Expressions: expressions}, nil
}

// Реализация метода GetExpressionByID
func (s *CalculatorServer) GetExpressionByID(ctx context.Context, req *pb.GetExpressionByIDRequest) (*pb.GetExpressionResponse, error) {
    login, err := validateTokenFromContext(ctx)
    if err != nil {
        return nil, status.Error(codes.Unauthenticated, err.Error())
    }

    // Получаем user_id из базы данных по логину
    var userID int
    err = s.DB.QueryRow("SELECT id FROM users WHERE login = $1", login).Scan(&userID)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, status.Error(codes.Unauthenticated, "User not found")
        }
        return nil, status.Error(codes.Internal, "Internal Server Error")
    }

    var expr pb.Expression
    err = s.DB.QueryRow("SELECT id, expression, result FROM calculations WHERE id = $1 AND user_id = $2", req.Id, userID).Scan(
        &expr.Id, &expr.Expression, &expr.Result,
    )
    if err == sql.ErrNoRows {
        return nil, status.Error(codes.NotFound, "Expression not found")
    } else if err != nil {
        return nil, status.Error(codes.Internal, "Internal Server Error")
    }

    return &pb.GetExpressionResponse{Expression: &expr}, nil
}


func (s *CalculatorServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
    if req.Login == "" || req.Password == "" {
        return nil, status.Error(codes.InvalidArgument, "логин и пароль обязательны")
    }

    var existingUser string
    err := s.DB.QueryRow(`SELECT login FROM users WHERE login = $1`, req.Login).Scan(&existingUser)
    if err == nil {
        return nil, status.Error(codes.AlreadyExists, "пользователь с таким логином уже существует")
    } else if !errors.Is(err, sql.ErrNoRows) {
        return nil, status.Error(codes.Internal, "ошибка базы данных")
    }

    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, status.Error(codes.Internal, "ошибка хеширования пароля")
    }

    _, err = s.DB.Exec(`INSERT INTO users (login, hashed_password) VALUES ($1, $2)`, req.Login, string(hashedPassword))
    if err != nil {
        return nil, status.Error(codes.Internal, "ошибка при сохранении пользователя: "+err.Error())
    }

    return &pb.RegisterResponse{
        Message: "пользователь успешно зарегистрирован",
    }, nil
}

func (s *CalculatorServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
    if req.Login == "" || req.Password == "" {
        return nil, status.Error(codes.InvalidArgument, "логин и пароль обязательны")
    }

    var hashedPassword string
    err := s.DB.QueryRow(`SELECT hashed_password FROM users WHERE login = $1`, req.Login).Scan(&hashedPassword)
    if errors.Is(err, sql.ErrNoRows) {
        return nil, status.Error(codes.Unauthenticated, "пользователь не найден")
    } else if err != nil {
        return nil, status.Error(codes.Internal, "ошибка базы данных")
    }

    err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password))
    if err != nil {
        return nil, status.Error(codes.Unauthenticated, "неверный пароль")
    }

    tokenString, err := generateJWTToken(req.Login)
    if err != nil {
        return nil, status.Error(codes.Internal, "ошибка создания токена")
    }

    return &pb.LoginResponse{
        Token: tokenString,
    }, nil
}



func validateTokenFromContext(ctx context.Context) (string, error) {
    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        return "", errors.New("metadata отсутствует")
    }

    authHeader := md["authorization"]
    if len(authHeader) == 0 {
        return "", errors.New("отсутствует заголовок authorization")
    }

    parts := strings.Split(authHeader[0], " ")
    if len(parts) != 2 || parts[0] != "Bearer" {
        return "", errors.New("неверный формат токена")
    }

    tokenString := parts[1]

    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("неподдерживаемый метод подписи")
        }
        return jwtKey, nil
    })

    if err != nil || !token.Valid {
        return "", errors.New("невалидный токен")
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return "", errors.New("не удалось извлечь claims")
    }

    login, ok := claims["login"].(string)
    if !ok {
        return "", errors.New("в токене отсутствует login")
    }

    return login, nil
}

func (s *CalculatorServer) DeleteAllTasks(ctx context.Context, req *pb.DeleteAllTasksRequest) (*pb.DeleteAllTasksResponse, error) {
    // Валидация токена для получения user_id (или просто использования токена для доступа)
    login, err := validateTokenFromContext(ctx)
    if err != nil {
        return nil, status.Error(codes.Unauthenticated, err.Error())
    }

    // Получаем user_id из базы данных по логину
    var userID int
    err = s.DB.QueryRow("SELECT id FROM users WHERE login = $1", login).Scan(&userID)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, status.Error(codes.Unauthenticated, "User not found")
        }
        return nil, status.Error(codes.Internal, "Internal Server Error")
    }

    // Удаляем все задачи пользователя из базы данных
    _, err = s.DB.Exec("DELETE FROM calculations WHERE user_id = $1", userID)
    if err != nil {
        log.Println("Ошибка при удалении задач:", err)
        return nil, status.Error(codes.Internal, "Internal Server Error")
    }

    return &pb.DeleteAllTasksResponse{
        Message: "Все задачи были удалены",
    }, nil
}