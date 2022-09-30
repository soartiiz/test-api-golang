// Lucas FOLLIOT
package todo

import (
	"context"
	"errors"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

type UserStorePG struct {
	db *pgxpool.Pool
}

func NewUserStorePG(db *pgxpool.Pool) *UserStorePG {
	_, err := db.Exec(context.Background(), `CREATE TABLE IF NOT EXISTS "user" ("id" TEXT PRIMARY KEY, "name" VARCHAR UNIQUE)`)

	if err != nil {
		panic(err)
	}

	return &UserStorePG{db}
}

func (pg *UserStorePG) Insert(user User) (*User, error) {
	var u User
	if _, err := pg.FindByName(user.Name); err == nil {
		return nil, errors.New("User already exist")
	}

	uuid := uuid.Must(uuid.NewGen().NewV4())
	insertStatement := `INSERT INTO "user" ("id", "name") VALUES ($1, $2) RETURNING *`
	response := pg.db.QueryRow(context.Background(), insertStatement, uuid, user.Name)

	err := response.Scan(&u.Id, &u.Name)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (pg *UserStorePG) FindById(id uuid.UUID) (*User, error) {
	var user User
	response := pg.db.QueryRow(context.Background(), `SELECT * FROM "user" WHERE id=$1`, id)

	err := response.Scan(&user.Id, &user.Name)
	if err != nil {
		panic(err)
	}

	return &user, nil
}

func (pg *UserStorePG) FindByName(name string) (*User, error) {
	var user User
	response := pg.db.QueryRow(context.Background(), `SELECT * FROM "user" WHERE name=$1`, name)

	err := response.Scan(&user.Id, &user.Name)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
