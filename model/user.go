// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// user.go [created: Sun, 28 Jul 2013]

package model

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Permission string

const (
	PermAdmin              = "ADMIN"
	PermMediaDelete        = "MEDIA_DELETE"
	PermMediaUpdate        = "MEDIA_UPDATE"
	PermUserList           = "USER_LIST"
	PermUserCreate         = "USER_CREATE"
	PermUserRead           = "USER_READ"
	PermUserUpdate         = "USER_UPDATE"
	PermUserDelete         = "USER_DELETE"
	PermUserProgressUpdate = "USER_PROGRESS_UPDATE"
)

func UserHasPermission(userid string, perm Permission) (bool, error) {
	q := `
		SELECT count(*)
		FROM UserPermissions AS up
		JOIN Users AS u ON u.UserId = up.UserId
		WHERE up.UserId = ? AND up.PermissionName = ?`
	row := DB.QueryRow(q, userid, string(perm))
	var count int
	err := row.Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

type User struct {
	Id      string    `json:"userId"`
	Email   string    `json:"-"`
	Created time.Time `json:"created"`
}

type accessToken struct {
	UserId string
	Token  string
}

func newAccessTokenForUser(userid string) (*accessToken, error) {
	at := new(accessToken)
	at.UserId = userid

	p := make([]byte, 16)
	_, err := rand.Read(p)
	if err != nil {
		return nil, err
	}
	at.Token = fmt.Sprintf("%x", p)
	return at, nil
}

func CreateAccessTokenForUser(userid string) (string, error) {
	at, err := newAccessTokenForUser(userid)
	if err != nil {
		return "", err
	}

	q := `INSERT INTO AccessTokens(UserId, AccessToken) Values(?, ?)`
	_, err = DB.Exec(q, at.UserId, at.Token)
	if err != nil {
		return "", err
	}

	return at.Token, nil
}

func AllAccessTokensForUser(userid string) ([]string, error) {
	accessTokens := make([]string, 0, 10)
	q := `SELECT AccessToken FROM AccessTokens WHERE UserId = ?`
	rows, err := DB.Query(q, userid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var token string
		err := rows.Scan(&token)
		if err != nil {
			return nil, err
		}
		accessTokens = append(accessTokens, token)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return accessTokens, nil
}

func FindUserByAccessToken(accessToken string) (*User, error) {
	u := new(User)
	q := `
		SELECT UserId, Email, Created
		FROM Users
		NATURAL JOIN UserEmails
		NATURAL JOIN AccessTokens
		WHERE AccessToken = ?`
	row := DB.QueryRow(q, strings.ToLower(accessToken))
	err := row.Scan(&u.Id, &u.Email, &u.Created)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// Find the the id of a user with the given email. If no user exists, one is created.
// The user's id is returned along with any error encountered.
func LocateOrCreateUserByEmail(email string) (string, error) {
	var id string
	q := `
		SELECT UserId
		FROM Users
		NATURAL JOIN UserEmails
		WHERE Email = ?`
	row := DB.QueryRow(q, email)
	err := row.Scan(&id)
	if err == sql.ErrNoRows {
		var tx *sql.Tx
		tx, err = DB.Begin()
		if err == nil {
			id = getsha1(email) // this may be problematic
			q := `INSERT INTO Users(UserId) VALUES (?)`
			_, err = tx.Exec(q, id)
			if err != nil {
				tx.Rollback()
			} else {
				q = `INSERT INTO UserEmails(UserId, Email) Values (?, ?)`
				_, err = tx.Exec(q, id, email)
				if err != nil {
					tx.Rollback()
				} else {
					tx.Commit()
				}
			}
		}
	}
	if err != nil {
		return "", err
	}
	return id, nil
}
