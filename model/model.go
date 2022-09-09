package model

import (
	"time"
)

type Todo struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	Complete  bool      `json:"completed"`
	CreatedAt time.Time `json:"created_at"`
}

type DBHandler interface {
	GetTodos(sessionId string) []*Todo
	AddTodos(sessionId string, name string) *Todo
	RemoveTodos(id int) bool
	CompleteTodos(id int, complete bool) bool
	Close()
}

func NewDBHandler(filepath string) DBHandler {
	return newSqliteHandler(filepath)
}
