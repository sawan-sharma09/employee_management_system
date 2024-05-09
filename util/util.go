package util

// Created this struct in a different package to avoid import cycle

type Employee struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Department string  `json:"department"`
	Salary     float64 `json:"salary"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
