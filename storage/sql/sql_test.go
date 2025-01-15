package sql

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/ilaziness/gokit/config"
	"github.com/stretchr/testify/assert"
)

var cfg = &config.DB{
	DSN: "root:root@tcp(127.0.0.1:3306)/ent_test",
}

type User struct {
	ID        int    `db:"id"`
	Age       int    `db:"age"`
	Name      string `db:"name"`
	Username  string `db:"username"`
	CreatedAt string `db:"created_at"`
}

func init() {
	InitSqlx(cfg)
}

func TestInitSqlx(t *testing.T) {
	u := User{}
	err := sqlxDB.Get(&u, "SELECT * FROM users LIMIT 1")
	assert.Equal(t, nil, err)
	assert.Greater(t, u.ID, 0)
	t.Log(u)
}
