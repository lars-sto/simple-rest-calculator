package model

import "time"

type Operation string

const (
	OpAdd Operation = "add"
	OpSub Operation = "subtract"
	OpMul Operation = "multiply"
	OpDiv Operation = "divide"
)

type ResultEntry struct {
	Time      time.Time `json:"time"`
	Op        Operation `json:"op"`
	A         float64   `json:"a"`
	B         float64   `json:"b"`
	Result    *float64  `json:"result,omitempty"`
	Error     string    `json:"error,omitempty"`
	ExprHuman string    `json:"exprHuman"`
}
