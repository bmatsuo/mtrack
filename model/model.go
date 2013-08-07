// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// model.go [created: Sat, 27 Jul 2013]

package model

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/bmatsuo/mtrack/model/migration"
	_ "github.com/mattn/go-sqlite3"
)

var DBPath string

var DB *sql.DB

type dbError struct {
	op  string
	err error
}

func (err dbError) Error() string {
	return fmt.Sprintf("%s: %v", err.op, err.err)
}

func DBDemoUser() error {
	userid, err := LocateOrCreateUserByEmail("bryan.matsuo@gmail.com")
	if err != nil {
		return dbError{"get user", err}
	}

	tokens, err := AllAccessTokensForUser(userid)
	if err != nil {
		return dbError{"get access tokens", err}
	}

	if len(tokens) == 0 {
		_, err := CreateAccessTokenForUser(userid)
		if err != nil {
			return dbError{"create access tokens", err}
		}
	}

	log.Print("ADMIN")
	log.Print(UserHasPermission(userid, "ADMIN"))

	return nil
}

func DBInit() error {
	var err error
	DB, err = sql.Open("sqlite3", DBPath)
	if err != nil {
		return dbError{"open", err}
	}
	err = dbInitMigrations()
	if err != nil {
		return dbError{"migrations", err}
	}
	return nil
}

func dbInitMigrations() error {
	err := migration.Initialize(DB)
	if err != nil {
		return err
	}

	Migrations := migration.MakeSequence(10)
	Migrations = Migrations.Append("001 create media table",
		migration.New(
			migration.String(
				`CREATE TABLE IF NOT EXISTS Media(
					MediaId  TEXT PRIMARY KEY ON CONFLICT ABORT,
					Root     TEXT NOT NULL,
					Path     TEXT NOT NULL,
					PathNorm TEXT NOT NULL,
					ModTime  DATETIME NOT NULL,
					Created  DATETIME DEFAULT CURRENT_TIMESTAMP
				)`,
			),
			migration.String(`DROP TABLE Media`),
		),
	)
	Migrations = Migrations.Append("002 create index MediaPathNorm",
		migration.New(
			migration.String(
				`CREATE INDEX IF NOT EXISTS MediaPathNorm ON Media (PathNorm ASC)`,
			),
			migration.String(`DROP INDEX MediaPathNorm`),
		),
	)
	Migrations = Migrations.Append("003 create index mediaroot",
		migration.New(
			migration.String(
				`CREATE INDEX IF NOT EXISTS MediaRoot ON Media (Root ASC)`,
			),
			migration.String(`DROP INDEX MediaRoot`),
		),
	)
	Migrations = Migrations.Append("004 create index mediamodtime",
		migration.New(
			migration.String(
				`CREATE INDEX IF NOT EXISTS MediaModTime ON Media (ModTime DESC)`,
			),
			migration.String(`DROP INDEX MediaModTime`),
		),
	)
	Migrations = Migrations.Append("005 create index mediacreated",
		migration.New(
			migration.String(
				`CREATE INDEX IF NOT EXISTS MediaCreated ON Media (Created DESC)`,
			),
			migration.String(`DROP INDEX MediaCreated`),
		),
	)
	Migrations = Migrations.Append("006 create table users",
		migration.New(
			migration.String(
				`CREATE TABLE IF NOT EXISTS Users(
					UserId	TEXT PRIMARY KEY,
					Created DATETIME DEFAULT CURRENT_TIMESTAMP
				)`,
			),
			migration.String(`DROP TABLE Users`),
		),
	)
	Migrations = Migrations.Append("007 create index userscreated",
		migration.New(
			migration.String(
				`CREATE INDEX IF NOT EXISTS UsersCreated ON Users (Created DESC)`,
			),
			migration.String(`DROP INDEX UsersCreated`),
		),
	)
	Migrations = Migrations.Append("010 create index useremails",
		migration.New(
			migration.String(
				`CREATE TABLE IF NOT EXISTS UserEmails(
					UserId	TEXT NOT NULL,
					Email   TEXT PRIMARY KEY ON CONFLICT ABORT,
					Created DATETIME DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (UserId) REFERENCES User(UserId)
				)`,
			),
			migration.String(`DROP TABLE UserEmails`),
		),
	)
	Migrations = Migrations.Append("011 create index accesstokens",
		migration.New(
			migration.String(
				`CREATE TABLE IF NOT EXISTS AccessTokens(
					UserId      TEXT NOT NULL,
					AccessToken TEXT PRIMARY KEY ON CONFLICT ABORT,
					FOREIGN KEY (UserId) REFERENCES Users(UserId)
				)`,
			),
			migration.String(`DROP TABLE AccessTokens`),
		),
	)
	Migrations = Migrations.Append("012 create index userstartedmedia",
		migration.New(
			migration.String(
				`CREATE TABLE IF NOT EXISTS UserStartedMedia(
					UserId  TEXT NOT NULL,
					MediaId TEXT NOT NULL,
					Started DATETIME DEFAULT CURRENT_TIMESTAMP,
					PRIMARY KEY (UserId, MediaId) ON CONFLICT ABORT,
					FOREIGN KEY (UserId) REFERENCES Users(UserId),
					FOREIGN KEY (MediaId) REFERENCES Media(MediaId)
				)`,
			),
			migration.String(`DROP TABLE UserStartedMedia`),
		),
	)
	Migrations = Migrations.Append("013 create index userstartedmediastarted",
		migration.New(
			migration.String(
				`CREATE INDEX IF NOT EXISTS UserStartedMediaStarted
					ON UserStartedMedia (Started DESC)`,
			),
			migration.String(`DROP INDEX UserStartedMediaStarted`),
		),
	)
	Migrations = Migrations.Append("014 create index userfinishedmedia",
		migration.New(
			migration.String(
				`CREATE TABLE IF NOT EXISTS UserFinishedMedia(
					UserId  TEXT NOT NULL,
					MediaId TEXT NOT NULL,
					Finished DATETIME DEFAULT CURRENT_TIMESTAMP,
					PRIMARY KEY (UserId, MediaId) ON CONFLICT ABORT,
					FOREIGN KEY (UserId) REFERENCES Users(UsersId),
					FOREIGN KEY (MediaId) REFERENCES Media(MediaId)
				)`,
			),
			migration.String(`DROP TABLE UserFinishedMedia`),
		),
	)
	Migrations = Migrations.Append("015 create index userfinishedmediafinished",
		migration.New(
			migration.String(
				`CREATE INDEX IF NOT EXISTS UserFinishedMediaFinished
					ON UserFinishedMedia (Finished DESC) `,
			),
			migration.String(`DROP INDEX UserFinishedMediaFinished`),
		),
	)
	Migrations = Migrations.Append("016 create index userpermissions",
		migration.New(
			migration.String(
				`CREATE TABLE IF NOT EXISTS UserPermissions(
					UserId TEXT NOT NULL,
					PermissionName TEXT NOT NULL,
					Created DATETIME DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (UserId) REFERENCES Users(UserId),
					PRIMARY KEY (UserId, PermissionName) ON CONFLICT IGNORE
				)`,
			),
			migration.String(`DROP TABLE UserPermissions`),
		),
	)

	ahead, behind, err := migration.Diff(DB, Migrations)
	switch {
	case err != nil:
		return err
	case ahead > 0:
		return fmt.Errorf("database is ahead of current migrations")
	case behind > 0:
		log.Printf("the database is %d migrations behind...", behind)
		n := Migrations.Len()
		_behind := Migrations.Slice(n-behind, n)
		log.Printf("running %d migrations...", _behind.Len())
		return migration.Apply(DB, migration.Up, -1, _behind)
	default:
		return nil
	}
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
