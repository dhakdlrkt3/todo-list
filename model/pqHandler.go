package model

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

type pqHandler struct {
	db *sql.DB
}

func (s *pqHandler) GetTodos(sessionId string) []*Todo {
	todos := []*Todo{}
	rows, err := s.db.Query("SELECT id, name, completed, createdAt FROM todos WHERE sessionId=$1", sessionId)
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

func (s *pqHandler) AddTodos(sessionId string, name string) *Todo {
	stmt, err := s.db.Prepare("INSERT INTO todos (sessionId, name, completed, createdAt) VALUES ($1, $2, $3, now()) RETURNING id")
	if err != nil {
		panic(err)
	}

	var id int
	err = stmt.QueryRow(sessionId, name, false).Scan(&id)
	if err != nil {
		panic(err)
	}

	var todo Todo
	todo.Id = id
	todo.Name = name
	todo.Complete = false
	todo.CreatedAt = time.Now()
	return &todo
}

func (s *pqHandler) RemoveTodos(id int) bool {
	stmt, err := s.db.Prepare("DELETE FROM todos WHERE id=$1")
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

func (s *pqHandler) CompleteTodos(id int, complete bool) bool {
	stmt, err := s.db.Prepare("UPDATE todos SET completed=$1 WHERE id=$2")
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

func (s pqHandler) Close() {
	s.db.Close()
}

func newPQHandler(dbConn string) DBHandler {
	database, err := sql.Open("postgres", dbConn)
	if err != nil {
		panic(err)
	}

	statement, err := database.Prepare(
		`CREATE TABLE IF NOT EXISTS todos (
			id        SERIAL  PRIMARY KEY,
			sessionId VARCHAR(256),
			name      TEXT,
			completed BOOLEAN,
			createdAt TIMESTAMP
		);
		`)

	if err != nil {
		panic(err)
	}

	_, err = statement.Exec()
	if err != nil {
		panic(err)
	}

	statement, err = database.Prepare(
		`CREATE INDEX IF NOT EXISTS sessionIdIndexOnTodos ON todos (
			sessionId ASC
		);
		`)

	if err != nil {
		panic(err)
	}

	_, err = statement.Exec()
	if err != nil {
		panic(err)
	}

	statement.Exec()

	return &pqHandler{
		db: database,
	}
}
