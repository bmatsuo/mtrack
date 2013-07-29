// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// media.go [created: Sun, 28 Jul 2013]

package model

import (
	"crypto/sha1"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"os"
	"strings"
)

type Media struct {
	Id      string    `json:"mediaId"`
	Path    string    `json:"path"`
	ModTime time.Time `json:"modified"`
}

var ErrNotImplemented = errors.New("not implemented")

func SyncMedia(path string, info os.FileInfo) (string, error) {
	pathnorm := strings.ToLower(path)
	sha1 := getsha1(pathnorm)

	var mod time.Time
	q := `select ModTime from Media where MediaId = ?`
	row := DB.QueryRow(q, sha1)
	err := row.Scan(&mod)
	switch err {
	case sql.ErrNoRows:
		q := `INSERT INTO Media(MediaId, Path, PathNorm, ModTime)`
		q += ` VALUES (?, ?, ?, ?)`
		_, err = DB.Exec(q, sha1, path, pathnorm, info.ModTime())
		if err != nil {
			return "", err
		}
	case nil:
		_mod := info.ModTime()
		if _mod.After(mod) {
			q := `UPDATE Media SET ModTime = ? WHERE MediaId = ?`
			_, err = DB.Exec(q, _mod, sha1)
			// want to return the existing id in this case
		}
	}
	return sha1, err
}

func getsha1(path string) string {
	h := sha1.New()
	h.Write([]byte(path))
	sum := h.Sum(nil)
	return fmt.Sprintf("%x", sum)
}

func FindMedia(id string) (*Media, error) {
	m := new(Media)
	q := `SELECT MediaId, Path, ModTime FROM Media WHERE MediaId = ?`
	row := DB.QueryRow(q, id)
	err := row.Scan(&m.Id, &m.Path, &m.ModTime)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func AllMedia() ([]*Media, error) {
	ms := make([]*Media, 0, 20)
	rows, err := DB.Query(`
		SELECT MediaId, Path, ModTime
		FROM Media
		ORDER BY ModTime DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		m := new(Media)
		err = rows.Scan(&m.Id, &m.Path, &m.ModTime)
		if err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return ms, nil
}
