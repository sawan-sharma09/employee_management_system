package util

// Created this struct in a different package to avoid import cycle

type Employee struct {
	ID         int     `json:"id" validate:"required,gt=0" `
	Name       string  `json:"name" validate:"required,min=2,max=30"`
	Department string  `json:"department" validate:"required"`
	Salary     float64 `json:"salary" validate:"required,gte=0"`
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
