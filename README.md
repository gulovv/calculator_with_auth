# Проект "Web Calculator"

Проект представляет собой распределённую систему вычисления арифметических выражений с использованием REST API и gRPC. Основная цель системы — приём арифметических выражений от пользователей, преобразование их в DAG (направленный ациклический граф), распределение задач вычисления между агентами и возврат результата. Авторизация и аутентификация пользователей реализована через gRPC, где пользователи проходят регистрацию и вход, получают JWT-токен и используют его для доступа к защищённым методам.

1. **Оркестратор (orchestrator)**
   
	Оркестратор играет центральную роль в управлении вычислениями. Он принимает арифметические выражения от клиентов, преобразует их в направленный ациклический граф (DAG), где каждая вершина представляет простую математическую операцию. Оркестратор управляет внутренней очередью задач, отслеживает их выполнение, рассылает подзадачи агентам и собирает результаты в итоговое значение выражения. Все данные сохраняются в базе данных.

3. **Агент (agent)**
   
   Агент — это исполняемый компонент, ожидающий задачи от оркестратора. Когда оркестратор отправляет агенту простую бинарную операцию (например, сложение или умножение двух чисел), агент выполняет её и возвращает результат. Такая архитектура позволяет масштабировать систему — чем больше агентов, тем быстрее обрабатываются сложные выражения, поскольку операции могут выполняться параллельно.

Если что-то не работает, пишите в Telegram: [gulovv](https://t.me/gulovv).

## Технологии

- **Go** — язык программирования для реализации сервисов.
- **Docker** — для контейнеризации сервисов.
- **Docker Compose** — для запуска нескольких сервисов с одной конфигурацией.

## Установка

Для установки проекта на вашем компьютере выполните следующие шаги.

**✅1. Клонируйте репозиторий:**


   ```bash
   git clone https://github.com/gulovv/calculator_with_auth.git
```

**✅2. Соберите и запустите образы Docker**:

   В директории проекта выполните команду:
   ```bash
   docker-compose up --build
```

**✅3. Запуск Docker контейнеров**

   После выполнения команды `docker-compose up --build` в терминале будет выведено следующее сообщение, это означает, что сервис собран и перезапущен:
   ```bash
   ✔ Service calculator              Built                                                                                                                                                     54.5s 
 ✔ Container calculator_db         Created                                                                                                                                                    0.0s 
 ✔ Container calculator_with_auth  Recreated                                                                                                                                                  0.7s 
Attaching to calculator_db, calculator_with_auth
calculator_db         | 
calculator_db         | PostgreSQL Database directory appears to contain a database; Skipping initialization
calculator_db         | 
calculator_db         | 2025-05-09 13:43:40.843 UTC [1] LOG:  starting PostgreSQL 14.18 (Debian 14.18-1.pgdg120+1) on aarch64-unknown-linux-gnu, compiled by gcc (Debian 12.2.0-14) 12.2.0, 64-bit
calculator_db         | 2025-05-09 13:43:40.852 UTC [1] LOG:  listening on IPv4 address "0.0.0.0", port 5432
calculator_db         | 2025-05-09 13:43:40.852 UTC [1] LOG:  listening on IPv6 address "::", port 5432
calculator_db         | 2025-05-09 13:43:40.855 UTC [1] LOG:  listening on Unix socket "/var/run/postgresql/.s.PGSQL.5432"
calculator_db         | 2025-05-09 13:43:40.862 UTC [27] LOG:  database system was shut down at 2025-05-09 13:42:42 UTC
calculator_db         | 2025-05-09 13:43:40.884 UTC [1] LOG:  database system is ready to accept connections
calculator_with_auth  | 2025/05/09 13:43:40 База данных уже существует, продолжаем
calculator_with_auth  | 2025/05/09 13:43:40 Таблица users проверена/создана
calculator_with_auth  | 2025/05/09 13:43:40 Таблица calculations проверена/создана
calculator_with_auth  | 2025/05/09 13:43:40 HTTP сервер запущен на :8080
calculator_with_auth  | 2025/05/09 13:43:40 gRPC сервер запущен на :50051
```

**✅4. Работа с проектом через терминал**

   Для взаимодействия с проектом можно использовать следующие инструменты:
   
•	REST API — с помощью curl
 
•	gRPC API — с помощью grpcurl

Рекомендуется открыть отдельный терминал для отправки запросов. Если что-то не работает, пишите в Telegram: [gulovv](https://t.me/gulovv).

## Общие ошибки, которые могут возникнуть в любом эндпоинте:

| Код ошибки         | Описание                                                                                                                                   |
|--------------------|--------------------------------------------------------------------------------------------------------------------------------------------|
| **400 Bad Request❌** | Этот код ошибки возникает, когда запрос не может быть обработан сервером из-за некорректных данных, отправленных в запросе. Ошибка может быть вызвана, например, если ID имеет неправильный формат (не число) или если выражение для вычисления содержит недопустимые символы. |
| **404 Not Found❌**   | Этот код ошибки возвращается, если запрашиваемый ресурс не найден на сервере. Это может произойти, если, например, задача с указанным ID не существует или был сделан запрос к несуществующему маршруту. |
| **500 Internal Server Error❌** | Этот код ошибки указывает на то, что произошла непредвиденная ошибка на сервере, из-за которой он не смог выполнить запрос. Обычно такая ошибка возникает при внутренних сбоях, например, при ошибке обработки данных, проблемах с подключением к базе данных или других сбоях в логике работы сервера. |

### Список доступных gRPC / HTTP методов

### ✅ 1. Регистрация пользователя

**POST /api/v1/register**
**gRPC: calculator.CalculatorService/Register**

Этот эндпоинт позволяет зарегистрировать нового пользователя. Для этого необходимо указать логин и пароль.

**Пример запроса через cURL:**

```bash
curl -X POST http://localhost:8080/api/v1/register \
    -H "Content-Type: application/json" \
    -d '{"login": "user1", "password": "password123"}'
```


**Пример запроса через gRPC:**

```bash
grpcurl -d '{"login": "user1", "password": "password123"}' \
    -plaintext localhost:50051 calculator.CalculatorService/Register
```

**Ответ:**

```bash
{
  "message": "пользователь успешно зарегистрирован"
}
```




⸻

### ✅ 2. Вход пользователя

**POST /api/v1/login**
**gRPC: calculator.CalculatorService/Login**

Этот эндпоинт позволяет пользователю войти в систему и получить JWT-токен для дальнейшего использования защищённых методов.

**Пример запроса через cURL:**

```bash
curl -X POST http://localhost:8080/api/v1/login \
    -H "Content-Type: application/json" \
    -d '{"login": "user1", "password": "password123"}'
```


**Пример запроса через gRPC:**

```bash
grpcurl -d '{"login": "user1", "password": "password123"}' \
    -plaintext localhost:50051 calculator.CalculatorService/Login
```

**Ответ**

```bash
{
  "token": "your_jwt_token_here"
}
```

Примечание: токен, полученный при входе, должен использоваться для доступа к защищённым методам API, добавляя его в заголовок Authorization в формате Bearer.

### ✅3. Вычисление арифметического выражения

**POST /api/v1/calculate**
**gRPC: calculator.CalculatorService/Calculate**

Позволяет отправить выражение на вычисление. Оркестратор разобьёт выражение на DAG и распределит подзадачи агентам.

**Пример запроса (HTTP):**
```
curl -X POST http://localhost:8080/api/v1/calculate \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer <ваш_токен>" \
    -d '{"expression": "2 * 9 + 8"}'
```
**Пример запроса (gRPC):**
```
grpcurl -d '{"expression": "2 * 9 + 8"}' \
    -H 'authorization: Bearer <ваш_токен>' \
    -plaintext localhost:50051 calculator.CalculatorService/Calculate
```
**Ответ:**
```
{
  "id": 1
}
```


⸻

### ✅4. Получение всех выражений

**GET /api/v1/expressions**
**gRPC: calculator.CalculatorService/GetExpressions**

Возвращает список всех выражений, отправленных текущим пользователем.

**Пример запроса (HTTP):**
```
curl -X GET http://localhost:8080/api/v1/expressions \
    -H "Authorization: Bearer <ваш_токен>"
```
**Пример запроса (gRPC):**
```
grpcurl -d '{}' \
    -H 'authorization: Bearer <ваш_токен>' \
    -plaintext localhost:50051 calculator.CalculatorService/GetExpressions
```
**Ответ:**
```
{
  "expressions": [
    {
      "id": 1,
      "expression": "2 * 9 + 8",
      "result": "26"
    }
  ]
}
```


⸻

### ✅5. Получение выражения по ID

**GET /api/v1/expressions/{id}**
**gRPC: calculator.CalculatorService/GetExpressionByID**

Позволяет получить конкретное выражение и его результат по ID.

**Пример запроса (HTTP):**

```
curl -X GET http://localhost:8080/api/v1/expressions/1 \
    -H "Authorization: Bearer <ваш_токен>"
```

**Пример запроса (gRPC):**

```
grpcurl -d '{"id": 1}' \
    -H 'authorization: Bearer <ваш_токен>' \
    -plaintext localhost:50051 calculator.CalculatorService/GetExpressionByID
```

**Ответ:**
```
{
  "id": 1,
  "expression": "2 * 9 + 8",
  "result": "26"
}
```


### ✅6. Удаление всех задач

**DELETE /api/v1/tasks/delete**
**gRPC: calculator.CalculatorService/DeleteAllTasks**

Удаляет все выражения и результаты текущего пользователя. Только для авторизованных пользователей.

**Пример запроса (HTTP):**
```
curl -X DELETE http://localhost:8080/api/v1/tasks/delete \
    -H "Authorization: Bearer <ваш_токен>"
```
**Пример запроса (gRPC):**
```
grpcurl -d '{}' \
    -H 'authorization: Bearer <ваш_токен>' \
    -plaintext localhost:50051 calculator.CalculatorService/DeleteAllTasks
```
**Ответ:**
```
{
  "message": "все задачи удалены"
}
```

## ⚠️ Примечание о различиях HTTP и gRPC ответов
Хотя методы регистрации и входа доступны как через HTTP, так и через gRPC, ответы имеют небольшие отличия в формате: 
| Отличие               | HTTP (curl)                                      | gRPC (grpcurl)                               |
|-----------------------|--------------------------------------------------|---------------------------------------------|
| Формат ответа      | Однострочный JSON                                | Красиво отформатированный JSON              |
| Формат ошибок      | JSON с полями code, message, details       | CLI-вывод: Code, Message                |
| Транспорт          | HTTP + JSON                                      | HTTP/2 + Protobuf (grpcurl отображает как JSON) |

## Структура
Здесь описана структура каталогов и файлов проекта web_calculator. Всё, что связано с основной логикой, обработкой запросов и вычислениями, упорядочено по соответствующим папкам.

```
├── cmd/
│   └── server/
│       └── main.go                # Точка входа: запуск HTTP и gRPC серверов
│
├── googleapis/                   # Протоколы и аннотации для gRPC-Gateway (google/api/*.proto)
│
├── internal/
│   ├── compute/
│   │   ├── agent.go              # Логика агента: выполнение простых операций
│   │   └── orchestrator.go      # Логика оркестратора: координация DAG и задач
│   │
│   ├── dag/
│   │   ├── builder.go            # Построение DAG из токенов
│   │   ├── rpn.go                # Обработка выражения в ОПЗ (обратная польская запись)
│   │   └── tokenizer.go         # Лексический анализатор (токенизация выражения)
│   │
│   ├── handler/
│   │   ├── grpc_server.go        # gRPC-сервер и обработчики
│   │   ├── http_server.go        # HTTP-сервер с gRPC-Gateway
│   │   ├── jwt.go                # Генерация и валидация JWT токенов
│   │   └── pb/                   # Сгенерированные gRPC и gRPC-Gateway файлы
│   │       ├── calculator.pb.go
│   │       ├── calculator_grpc.pb.go
│   │       └── calculator.pb.gw.go
│
├── pkg/
│   └── model/
│       └── model.go              # Общие модели данных (Expression, Task и др.)
│
├── proto/
│   └── calculator.proto          # gRPC-протокол: описание всех методов и сообщений
│
├── storage/
│   └── storage.go                # Работа с базой данных (PostgreSQL)
│
├── test/
│   └── handler_test.go           # Тесты для gRPC и HTTP обработчиков
│
└── README.md                     # Документация проекта
```
## Связь со мной

Если возникили какие-то вопросы, можете писать в Telegram: [gulovv](https://t.me/gulovv).


