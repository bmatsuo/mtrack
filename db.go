// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// db.go [created: Sat, 27 Jul 2013]

package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var DBPath string

var DB *sql.DB

func DBInit() error {
	var err error
	DB, err = sql.Open("sqlite3", DBPath)
	if err != nil {
		return err
	}
	return dbInitMigrations()
}

func dbInitMigrations() error {
	return dbQueryChain(
		`
		CREATE TABLE IF NOT EXISTS Media(
			MediaId TEXT PRIMARY KEY ON CONFLICT ABORT,
			Path TEXT NOT NULL,
			ModTime DATETIME NOT NULL,
			Created DATETIME DEFAULT CURRENT_TIMESTAMP
		)
		`,
		`
		CREATE TABLE IF NOT EXISTS Users(
			Email   TEXT PRIMARY KEY,
			Created DATETIME DEFAULT CURRENT_TIMESTAMP
		)
		`,
		`
		CREATE TABLE IF NOT EXISTS UserStartedMedia(
			Email   TEXT,
			MediaId TEXT,
			Started DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (Email, MediaId) ON CONFLICT ABORT,
			FOREIGN KEY (Email) REFERENCES Users(Email),
			FOREIGN KEY (MediaId) REFERENCES Media(Id)
		)
		`,
		`
		CREATE TABLE IF NOT EXISTS UserFinishedMedia(
			Email    TEXT,
			MediaId  TEXT,
			Finished DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (Email, MediaId) ON CONFLICT ABORT,
			FOREIGN KEY (Email) REFERENCES Users(Email),
			FOREIGN KEY (MediaId) REFERENCES Media(Id)
		)
		`,
	)
}

func dbQueryChain(query ...string) error {
	for _, q := range query {
		_, err := DB.Exec(q)
		if err != nil {
			return err
		}
	}
	return nil
}
