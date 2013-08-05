// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// progress.go [created: Sun, 28 Jul 2013]

package model

import (
	"database/sql"
	"errors"
	"time"
)

var ErrAlreadyStarted = errors.New("already started")
var ErrAlreadyFinished = errors.New("already finished")

type ActionStarted struct {
	MediaId   string    `json:"mediaId"`
	UserId    string    `json:"userId"`
	StartTime time.Time `json:"started"`
}

type ActionFinished struct {
	MediaId    string    `json:"mediaId"`
	UserId     string    `json:"userId"`
	FinishTime time.Time `json:"finished"`
}

func AllInProgress() ([]*ActionStarted, error) {
	as := make([]*ActionStarted, 0, 20)
	rows, err := DB.Query(`
		SELECT S.MediaId, S.UserId, S.Started
		FROM UserStartedMedia as S
		NATURAL JOIN Media
		ORDER BY Started DESC
	`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		a := new(ActionStarted)
		err = rows.Scan(&a.MediaId, &a.UserId, &a.StartTime)
		if err != nil {
			return nil, err
		}
		as = append(as, a)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return as, nil
}

func AllFinished() ([]*ActionFinished, error) {
	as := make([]*ActionFinished, 0, 20)
	rows, err := DB.Query(`
		SELECT S.MediaId, S.UserId, S.Finished
		FROM UserFinishedMedia as S
		NATURAL JOIN Media
		ORDER BY S.Finished DESC
	`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		a := new(ActionFinished)
		err = rows.Scan(&a.MediaId, &a.UserId, &a.FinishTime)
		if err != nil {
			return nil, err
		}
		as = append(as, a)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return as, nil
}

func ClearProgress(userid, mediaid string) error {
	q := `DELETE FROM UserStartedMedia WHERE MediaId = ? AND UserId = ?`
	_, err := DB.Exec(q, mediaid, userid)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	q = `DELETE FROM UserFinishedMedia WHERE MediaId = ? AND UserId = ?`
	_, err = DB.Exec(q, mediaid, userid)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	return nil
}

func StartMedia(userid, mediaid string) error {
	row := DB.QueryRow(`
		SELECT COUNT(*)
		FROM UserStartedMedia
		WHERE MediaId = ? AND UserId = ?`,
		mediaid, userid)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrAlreadyStarted
	}

	row = DB.QueryRow(`
		SELECT COUNT(*)
		FROM UserFinishedMedia
		WHERE MediaId = ? AND UserId = ?`,
		mediaid, userid)
	err = row.Scan(&count)
	if err != nil {
		return err
	}
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	if count > 0 {
		q := `DELETE FROM UserFinishedMedia WHERE MediaId = ? AND UserId = ?`
		_, err := tx.Exec(q, mediaid, userid)
		if err != nil {
			err = tx.Rollback()
			return err
		}
	}

	q := `INSERT INTO UserStartedMedia(MediaId, UserId) Values(?, ?)`
	_, err = tx.Exec(q, mediaid, userid)
	if err != nil {
		err := tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func FinishMedia(userid, mediaid string) error {
	row := DB.QueryRow(`
		SELECT COUNT(*)
		FROM UserFinishedMedia
		WHERE MediaId = ? AND UserId = ?`,
		mediaid, userid)
	var count int
	err := row.Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrAlreadyStarted
	}

	row = DB.QueryRow(`
		SELECT COUNT(*)
		FROM UserStartedMedia
		WHERE MediaId = ? AND UserId = ?`,
		mediaid, userid)
	err = row.Scan(&count)
	if err != nil {
		return err
	}
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	if count > 0 {
		q := `DELETE FROM UserStartedMedia WHERE MediaId = ? AND UserId = ?`
		_, err := tx.Exec(q, mediaid, userid)
		if err != nil {
			err = tx.Rollback()
			return err
		}
	}

	q := `INSERT INTO UserFinishedMedia(MediaId, UserId) Values(?, ?)`
	_, err = tx.Exec(q, mediaid, userid)
	if err != nil {
		err := tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
