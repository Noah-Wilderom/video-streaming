package migrations

import (
	"fmt"
	"gofr.dev/pkg/gofr/migration"
)

const createTable = `CREATE TABLE users (
  id VARCHAR(255) PRIMARY KEY KEY NOT NULL,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW()
);`

func createTableUsers() migration.Migrate {
	return migration.Migrate{
		UP: func(d migration.Datasource) error {
			fmt.Println("creating users table")
			_, err := d.SQL.Exec(createTable)
			if err != nil {
				fmt.Printf("users migration error: %s\n", err.Error())
				return err
			}

			return nil
		},
	}
}
