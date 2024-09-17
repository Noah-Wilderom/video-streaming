package migrations

import (
	"fmt"
	"gofr.dev/pkg/gofr/migration"
)

const createTable = `CREATE TABLE video_sessions (
  id VARCHAR(255) PRIMARY KEY KEY NOT NULL,
  user_id VARCHAR(255) NOT NULL,
  video_id VARCHAR(255) NOT NULL,
  fragment_hash VARCHAR(255) NOT NULL,
  fragment_path VARCHAR(255) NOT NULL,
  token VARCHAR(255) NOT NULL UNIQUE,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW()
);`

func createTableVideoSessions() migration.Migrate {
	return migration.Migrate{
		UP: func(d migration.Datasource) error {
			fmt.Println("creating video_sessions table")
			_, err := d.SQL.Exec(createTable)
			if err != nil {
				fmt.Printf("video_sessions migration error: %s\n", err.Error())
				return err
			}

			return nil
		},
	}
}
