// Lucas FOLLIOT
package todo

import (
	"context"
	"errors"
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

type TodoStorePG struct {
	db *pgxpool.Pool
}

func NewTodoStorePG(db *pgxpool.Pool) *TodoStorePG {
	_, err := db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS "todo" (
			"id" TEXT PRIMARY KEY,
			"text" VARCHAR,
			"done" BOOL,
			"user_id" TEXT
		)
	`)

	if err != nil {
		panic(err)
	}

	return &TodoStorePG{db}
}

func (pg *TodoStorePG) Add(todo Todo) (*Todo, error) {
	uuid := uuid.Must(uuid.NewGen().NewV4())
	insertStatement := `INSERT INTO "todo" ("id", "text", "done", "user_id") VALUES ($1, $2, $3, $4) RETURNING *`
	response := pg.db.QueryRow(context.Background(), insertStatement, uuid, todo.Text, todo.Done, todo.UserId)

	err := response.Scan(&todo.Id, &todo.Text, &todo.Done, &todo.UserId)
	if err != nil {
		return nil, err
	}

	return &todo, nil
}

func (pg *TodoStorePG) Delete(todoID uuid.UUID) error {
	response := pg.db.QueryRow(context.Background(), `DELETE FROM "todo" WHERE id = '`+todoID.String()+`';`)

	if response != nil {
		return errors.New("Error when deleting")
	}

	return nil
}

func (pg *TodoStorePG) Toggle(todoID uuid.UUID, done bool) (*Todo, error) {
	return nil, nil
}

func (pg *TodoStorePG) UpdateText(todoID uuid.UUID, text string) (*Todo, error) {
	var todo Todo
	response := pg.db.QueryRow(context.Background(), `UPDATE "todo" SET text = '`+text+`' WHERE id = '`+todoID.String()+`' RETURNING *;`)

	err := response.Scan(&todo.Id, &todo.Text, &todo.Done, &todo.UserId)
	if err != nil {
		return nil, err
	}

	return &todo, nil
}

func (pg *TodoStorePG) FindByID(todoID uuid.UUID) (*Todo, error) {
	var todo Todo
	response := pg.db.QueryRow(context.Background(), `SELECT * FROM "todo" WHERE id = '`+todoID.String()+"';")

	err := response.Scan(&todo.Id, &todo.Text, &todo.Done, &todo.UserId)
	if err != nil {
		return nil, err
	}

	return &todo, nil
}

func (pg *TodoStorePG) FindByUserID(userID uuid.UUID) ([]*Todo, error) {
	var todos []*Todo

	rows, erro := pg.db.Query(context.Background(), `SELECT * FROM "todo" WHERE user_id = '`+userID.String()+"';")
	if erro != nil {
		return nil, erro
	}

	defer rows.Close()

	for rows.Next() {
		var todo Todo

		rows.Scan(&todo.Id, &todo.Text, &todo.Done, &todo.UserId)

		todos = append(todos, &todo)
	}

	fmt.Println(todos)

	// err := response.Scan(&todo.Id, &todo.Text, &todo.Done, &todo.UserId)
	// err := response.Scan()
	// if err != nil {
	// 	return nil, err
	// }

	return todos, nil
}
