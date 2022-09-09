package model

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type sqliteHandler struct {
	db *sql.DB
}

func (s *sqliteHandler) GetTodos(sessionId string) []*Todo {
	todos := []*Todo{}
	rows, err := s.db.Query("SELECT id, name, completed, createdAt FROM todos WHERE sessionId=?", sessionId)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var todo Todo
		rows.Scan(&todo.Id, &todo.Name, &todo.Complete, &todo.CreatedAt)
		todos = append(todos, &todo)
	}

	return todos
}

func (s *sqliteHandler) AddTodos(sessionId string, name string) *Todo {
	stmt, err := s.db.Prepare("INSERT INTO todos (sessionId, name, completed, createdAt) VALUES (?, ?, ?, datetime('now'))")
	if err != nil {
		panic(err)
	}
	rst, err := stmt.Exec(sessionId, name, false)
	if err != nil {
		panic(err)
	}

	id, _ := rst.LastInsertId()
	var todo Todo
	todo.Id = int(id)
	todo.Name = name
	todo.Complete = false
	todo.CreatedAt = time.Now()
	return &todo
}

func (s *sqliteHandler) RemoveTodos(id int) bool {
	stmt, err := s.db.Prepare("DELETE FROM todos WHERE id=?")
	if err != nil {
		panic(err)
	}
	rst, err := stmt.Exec(id)
	if err != nil {
		panic(err)
	}

	cnt, _ := rst.RowsAffected()

	return cnt > 0
}

func (s *sqliteHandler) CompleteTodos(id int, complete bool) bool {
	stmt, err := s.db.Prepare("UPDATE todos SET completed=? WHERE id=?")
	if err != nil {
		panic(err)
	}

	rst, err := stmt.Exec(complete, id)

	if err != nil {
		panic(err)
	}

	cnt, _ := rst.RowsAffected()

	return cnt > 0
}

func (s sqliteHandler) Close() {
	s.db.Close()
}

func newSqliteHandler(filepath string) DBHandler {
	database, err := sql.Open("sqlite3", filepath)
	if err != nil {
		panic(err)
	}

	statement, _ := database.Prepare(
		`CREATE TABLE IF NOT EXISTS todos (
			id        INTEGER  PRIMARY KEY AUTOINCREMENT,
			sessionId STRING,
			name      TEXT,
			completed BOOLEAN,
			createdAt DATETIME
		);
		CREATE INDEX IF NOT EXISTS sessionIdIndexOnTodos ON todos (
			sessionId ASC
		);
		`)

	statement.Exec()

	return &sqliteHandler{
		db: database,
	}
}
