// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// user.go [created: Sun, 28 Jul 2013]

package model

import (
	"database/sql"
	"time"
)

type User struct {
	Id      string    `json:"userId"`
	Email   string    `json:"-"`
	Created time.Time `json:"created"`
}

func FindOrCreateUserByEmail(email string) (string, error) {
	var id string
	q := `SELECT UserId FROM Users WHERE Email = ?`
	row := DB.QueryRow(q, email)
	err := row.Scan(&id)
	if err == sql.ErrNoRows {
		id = getsha1(email)
		q := `INSERT INTO Users(UserId, Email) VALUES (?, ?)`
		_, err = DB.Exec(q, id, email)
	}
	if err != nil {
		return "", err
	}
	return id, nil
}
