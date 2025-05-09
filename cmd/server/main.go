package main

import (
    "context"
    "log"
    "net"
    "net/http"

    "github.com/gulovv/calculator_with_auth/internal/handler"
    "github.com/gulovv/calculator_with_auth/storage"
    "github.com/joho/godotenv"
    "google.golang.org/grpc"
    "google.golang.org/grpc/reflection"
    "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
    pb "github.com/gulovv/calculator_with_auth/internal/handler/pb"
)

func init() {
    // Загружаем переменные окружения из файла .env
    if err := godotenv.Load(); err != nil {
        log.Fatal("Error loading .env file")
    }
}

func main() {
    // Инициализируем базу данных
    db, err := storage.InitDB()
    if err != nil {
        log.Fatal("Ошибка инициализации базы данных:", err)
    }
    defer db.Close()

    // Настройка gRPC сервера
    grpcServer := grpc.NewServer()

    // Инициализируем gRPC-сервер с базой данных
    grpcHandler := &handler.CalculatorServer{
        DB: db,
    }

    // Регистрируем сервисы gRPC
    pb.RegisterCalculatorServiceServer(grpcServer, grpcHandler)

    // Reflection для gRPC
    reflection.Register(grpcServer)

    // Запуск gRPC сервера
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatal("Ошибка при запуске listener:", err)
    }
    go func() {
        log.Println("gRPC сервер запущен на :50051")
        log.Fatal(grpcServer.Serve(lis))
    }()

    // Настройка HTTP Gateway
    gwMux := runtime.NewServeMux()
    // Регистрируем gRPC обработчики для HTTP
    err = pb.RegisterCalculatorServiceHandlerServer(context.Background(), gwMux, grpcHandler)
    if err != nil {
        log.Fatalf("Ошибка регистрации gRPC-Gateway обработчиков: %v", err)
    }

    // HTTP сервер
    http.Handle("/", gwMux)

    // Запуск HTTP сервера
    log.Println("HTTP сервер запущен на :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

