// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// sync.go [created: Tue,  6 Aug 2013]

package migration

import (
	"database/sql"
	"fmt"
)

var (
	_CREATE = `
		CREATE TABLE IF NOT EXISTS Migrations(
			Name VARCHAR[255] PRIMARY KEY
		)
	`
	_SELECT = `
		SELECT *
		FROM Migrations
		ORDER BY Name ASC
	`
)

func Initialize(db *sql.DB) error {
	_, err := db.Exec(_CREATE)
	return err
}

func Applied(db *sql.DB) ([]string, error) {
	rows, err := db.Query(_CREATE)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, 10)
	for rows.Next() {
		var name string
		err = rows.Scan(&name)
		if err != nil {
			return nil, err
		}
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return names, nil
}

type Direction uint

const (
	Up Direction = iota
	Down
)

func Apply(db *sql.DB, dir Direction, n int, ms ...Interface) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	var bounds func() (int, int)
	var next func(int) int
	var migrate func(m Interface) error
	switch dir {
	case Up:
		bounds = func() (int, int) { return 0, len(ms) }
		next = func(i int) int { return i + 1 }
		migrate = func(m Interface) error { return m.Up(tx) }
	case Down:
		bounds = func() (int, int) { return len(ms) - 1, -1 }
		next = func(i int) int { return i - 1 }
		migrate = func(m Interface) error { return m.Down(tx) }
	default:
		return fmt.Errorf("unrecognized Direction: %d", dir)
	}

	rollbackError := func(err error) error {
		if err != nil {
			_err := tx.Rollback()
			if _err != nil {
				return fmt.Errorf("couldn't rollback: %v", _err)
			}
		}
		return err
	}

	count := 0
	for i, bound := bounds(); i != bound; i = next(i) {
		if n > 0 && count >= n {
			break
		}
		err = rollbackError(migrate(ms[i]))
		if err != nil {
			return err
		}
	}

	return rollbackError(tx.Commit())
}

var ErrOutOfOrder = fmt.Errorf("out of order migration")

func Diff(db *sql.DB, seq Sequence) (int, int, error) {
	names, err := Applied(db)
	if err != nil {
		return 0, 0, err
	}
	ahead, behind := len(names), seq.Len()
	for i := range names {
		if behind == 0 {
			break
		}
		name, _ := seq.Index(i)
		if name != names[i] {
			break
		}
		ahead--
		behind--
	}
	if ahead > 0 && behind > 0 {
		return ahead, behind, ErrOutOfOrder
	}
	return ahead, behind, nil
}

type Sequence struct {
	ms  []*namedMigration
	err error
}

func MakeSequence(capacity int) Sequence {
	var seq Sequence
	if capacity > 0 {
		seq.ms = make([]*namedMigration, 0, capacity)
	}
	return seq
}

func (seq Sequence) Len() int {
	return len(seq.ms)
}

// a panic occurs if i is out of range.
func (seq Sequence) Index(i int) (name string, m Interface) {
	n := seq.ms[i]
	return n.Name, n.M
}

func (seq Sequence) Append(name string, m Interface) Sequence {
	if seq.err != nil {
		return seq
	}
	for i := range seq.ms {
		if name == seq.ms[i].Name {
			seq.err = fmt.Errorf("duplicate migration: %v")
			return seq
		}
	}
	seq.ms = append(seq.ms, &namedMigration{name, m})
	return seq
}
