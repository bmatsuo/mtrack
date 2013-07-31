// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// user.go [created: Sun, 28 Jul 2013]

package model

import (
	"crypto/rand"
	"fmt"
	"database/sql"
	"strings"
	"time"
)

type Permission string

func UserHasPermission(userid string, perm Permission) (bool, error) {
	fmt.Println(userid, perm)
	q := `
		SELECT count(*)
		FROM UserPermissions AS up
		JOIN Users AS u ON u.UserId = up.UserId
		JOIN Permissions AS p ON p.PermissionName = up.PermissionName
		WHERE up.UserId = ? AND (
			up.PermissionName = 'ADMIN' OR
			up.PermissionName = ?
		)`
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
		NATURAL JOIN AccessTokens
		WHERE AccessToken = ?`
	row := DB.QueryRow(q, strings.ToLower(accessToken))
	err := row.Scan(&u.Id, &u.Email, &u.Created)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func LocateOrCreateUserByEmail(email string) (string, error) {
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
