// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// model.go [created: Sat, 27 Jul 2013]

package model

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
	err = dbInitMigrations()
	if err != nil {
		return err
	}

	userid, err := LocateOrCreateUserByEmail("bryan.matsuo@gmail.com")
	if err != nil {
		return err
	}

	tokens, err := AllAccessTokensForUser(userid)
	if err != nil {
		return err
	}

	if len(tokens) == 0 {
		_, err := CreateAccessTokenForUser(userid)
		if err != nil {
			return err
		}
	}

	return nil
}

func dbInitMigrations() error {
	return dbQueryChain(
		// media
		`CREATE TABLE IF NOT EXISTS Media(
			MediaId  TEXT PRIMARY KEY ON CONFLICT ABORT,
			Path     TEXT NOT NULL,
			PathNorm TEXT NOT NULL,
			ModTime  DATETIME NOT NULL,
			Created  DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS MediaPathNorm ON Media (PathNorm ASC)`,
		`CREATE INDEX IF NOT EXISTS MediaModTime ON Media (ModTime DESC)`,
		`CREATE INDEX IF NOT EXISTS MediaCreated ON Media (Created DESC)`,
		// users
		`CREATE TABLE IF NOT EXISTS Users(
			UserId	TEXT PRIMARY KEY,
			Email   TEXT UNIQUE ON CONFLICT ABORT,
			Created DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS UsersCreated ON Users (Created DESC)`,
		`CREATE INDEX IF NOT EXISTS UsersEmail ON Users (Email ASC)`,
		// access tokens
		`CREATE TABLE IF NOT EXISTS AccessTokens(
			UserId      TEXT NOT NULL,
			AccessToken TEXT PRIMARY KEY ON CONFLICT ABORT,
			FOREIGN KEY (UserId) REFERENCES Users(Id)
		)`,
		// a user started consuming media
		`CREATE TABLE IF NOT EXISTS UserStartedMedia(
			UserId  TEXT NOT NULL,
			MediaId TEXT NOT NULL,
			Started DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (UserId, MediaId) ON CONFLICT ABORT,
			FOREIGN KEY (UserId) REFERENCES Users(Id),
			FOREIGN KEY (MediaId) REFERENCES Media(Id)
		)`,
		`CREATE INDEX IF NOT EXISTS UserStartedMediaStarted
			ON UserStartedMedia (Started DESC)`,
		// a user finished consuming media
		`CREATE TABLE IF NOT EXISTS UserFinishedMedia(
			UserId  TEXT NOT NULL,
			MediaId TEXT NOT NULL,
			Finished DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (UserId, MediaId) ON CONFLICT ABORT,
			FOREIGN KEY (UserId) REFERENCES Users(Id),
			FOREIGN KEY (MediaId) REFERENCES Media(Id)
		)`,
		`CREATE INDEX IF NOT EXISTS UserFinishedMediaFinished
			ON UserFinishedMedia (Finished DESC) `,
		// permissions
		`CREATE TABLE IF NOT EXISTS Permissions(
			PermissionName TEXT PRIMARY KEY ON CONFLICT IGNORE
		)`,
		`INSERT INTO Permissions(PermissionName) Values('USER_LIST')`,
		`INSERT INTO Permissions(PermissionName) Values('USER_CREATE')`,
		`INSERT INTO Permissions(PermissionName) Values('USER_READ')`,
		`INSERT INTO Permissions(PermissionName) Values('USER_UPDATE')`,
		`INSERT INTO Permissions(PermissionName) Values('USER_DELETE')`,
		`INSERT INTO Permissions(PermissionName) Values('MEDIA_DELETE')`,
		`INSERT INTO Permissions(PermissionName) Values('MEDIA_UPDATE')`,
		`CREATE TABLE IF NOT EXISTS UserPermissions(
			UserId TEXT NOT NULL,
			PermissionName TEXT NOT NULL,
			FOREIGN KEY (UserId) REFERENCES User(Id),
			FOREIGN KEY (PermissionName) REFERENCES Permissions(PermissionName),
			PRIMARY KEY (UserId, PermissionName) ON CONFLICT IGNORE
		)`,
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
