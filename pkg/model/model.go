package model
// Для внутренней логики
type Node struct {
	ID      int
	Op      string
	LeftID  int
	RightID int
}

type Result struct {
	ID    int
	Value float64
}


// Запросы по пути localhost/api/v1/calculate
type Request struct {
	Expression string `json:"expression"`
}

type Response struct {
	Result float64 `json:"result"`
	Error  string  `json:"error,omitempty"`
}


// Запросы по пути localhost/api/v1/expressions и localhost/api/v1/expressions/{id}
type ExpressionResponse struct {
    ID         string  `json:"id"`
    Expression string  `json:"expression"`
    Result     float64 `json:"result"`
}

type ExpressionsListResponse struct {
    Expressions []ExpressionResponse `json:"expressions"`
}


// Запросы по пути localhost/api/v1/register и localhost/api/v1/login
type RegisterRequest struct {
    Login    string `json:"login"`
    Password string `json:"password"`
}

type LoginRequest struct {
    Login    string `json:"login"`
    Password string `json:"password"`
}
