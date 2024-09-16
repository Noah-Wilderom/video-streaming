package migrations

import (
	"fmt"
	"gofr.dev/pkg/gofr/migration"
)

const createTable = `CREATE TABLE videos (
  id VARCHAR(255) PRIMARY KEY KEY NOT NULL,
  user_id VARCHAR(255) NOT NULL,
  status VARCHAR(255) NOT NULL,
  path VARCHAR(255) NOT NULL,
  size INT NOT NULL,
  mimetype VARCHAR(255) NOT NULL,
  metadata TEXT NOT NULL,
  processed_at TIMESTAMP DEFAULT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW()
);`

func createTableVideos() migration.Migrate {
	return migration.Migrate{
		UP: func(d migration.Datasource) error {
			fmt.Println("creating videos table")
			_, err := d.SQL.Exec(createTable)
			if err != nil {
				fmt.Printf("videos migration error: %s\n", err.Error())
				return err
			}

			return nil
		},
	}
}
