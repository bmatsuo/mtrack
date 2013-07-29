// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// progress.go [created: Sun, 28 Jul 2013]

package model

import (
	"errors"
	"time"
)

var ErrAlreadyStarted = errors.New("already started")
var ErrAlreadyFinished = errors.New("already finished")

type ActionStarted struct {
	MediaId   string
	UserId    string
	StartTime time.Time
}

type ActionFinished struct {
	MediaId    string
	UserId     string
	FinishTime time.Time
}

func AllInProgressMedia(UserId string) ([]*Media, error) {
	return nil, nil
}

func AllFinishedMedia(UserId string) ([]*Media, error) {
	return nil, nil
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
