// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// sync.go [created: Tue,  6 Aug 2013]

package migration

import (
	"database/sql"
	"fmt"
	"sort"
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
	rows, err := db.Query(_SELECT)
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
		names = append(names, name)
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

func Apply(db *sql.DB, dir Direction, n int, seq Sequence) error {
	if seq.err != nil {
		return fmt.Errorf("sequence error: %v", seq.err)
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	var bounds func() (int, int)
	var next func(int) int
	var migrate func(string, Interface) error
	switch dir {
	case Up:
		bounds = func() (int, int) { return 0, seq.Len() }
		next = func(i int) int { return i + 1 }
		migrate = func(name string, m Interface) error {
			_, err := tx.Exec(`INSERT INTO Migrations (Name) Values (?)`, name)
			if err != nil {
				return err
			}
			return m.Up(tx)
		}
	case Down:
		bounds = func() (int, int) { return seq.Len() - 1, -1 }
		next = func(i int) int { return i - 1 }
		migrate = func(name string, m Interface) error {
			_, err := tx.Exec(`DELETE FROM Migrations WHERE Name = ?`, name)
			if err != nil {
				return err
			}
			return m.Down(tx)
		}
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
		if n >= 0 && count >= n {
			break
		}
		err = rollbackError(migrate(seq.Index(i)))
		if err != nil {
			return err
		}
	}

	return rollbackError(tx.Commit())
}

var ErrOutOfOrder = fmt.Errorf("out of order migration")

func Diff(db *sql.DB, seq Sequence) (int, int, error) {
	if seq.err != nil {
		return 0, 0, fmt.Errorf("sequence error: %v", seq.err)
	}
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

/*
a sequence of migrations. analogous to a slice of migrations.
*/
type Sequence struct {
	ms  []*namedMigration
	err error
}

// create a migration sequence preallocated with a specified capacity.
// no memory is allocated if capacity is less than or equal to zero.
func MakeSequence(capacity int) Sequence {
	var seq Sequence
	if capacity > 0 {
		seq.ms = make([]*namedMigration, 0, capacity)
	}
	return seq
}

func (seq Sequence) Err() error {
	return seq.err
}

// a panic occurs if i is out of range.
func (seq Sequence) Index(i int) (name string, m Interface) {
	n := seq.ms[i]
	return n.Name, n.M
}

func (seq Sequence) Len() int {
	return len(seq.ms)
}

func (seq Sequence) Less(i, j int) bool {
	return seq.ms[i].Name < seq.ms[j].Name
}

func (seq Sequence) Swap(i, j int) {
	seq.ms[i], seq.ms[j] = seq.ms[j], seq.ms[i]
}

func (seq Sequence) Sort() {
	sort.Sort(seq)
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

func (seq Sequence) Slice(i, j int) Sequence {
	seq.ms = seq.ms[i:j]
	return seq
}
