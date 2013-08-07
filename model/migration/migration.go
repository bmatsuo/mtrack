// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// migration.go [created: Tue,  6 Aug 2013]

// Package migration does migrations on sqlite3 databases.
// Specifically sqlite3 because I want simplicity and ease of use over
// reusability in this initial draft.
package migration

import (
	"database/sql"
	"errors"
)

var ErrIrreversible = errors.New("irreversible migration")

type Interface interface {
	Up(db *sql.Tx) error
	Down(db *sql.Tx) error
}

type Executor interface {
	Exec(db *sql.Tx) error
}

type ExecutorFunc func(db *sql.Tx) error

func (fn ExecutorFunc) Exec(db *sql.Tx) error {
	return fn(db)
}

var executorIrreversible = ExecutorFunc(func(db *sql.Tx) error {
	return ErrIrreversible
})

func MigrationIrreversible(up Executor) Interface {
	return New(up, executorIrreversible)
}

type simpleMigration struct {
	up, down Executor
}

func (m *simpleMigration) Up(db *sql.Tx) error {
	return m.up.Exec(db)
}

func (m *simpleMigration) Down(db *sql.Tx) error {
	return m.down.Exec(db)
}

func New(up, down Executor) Interface {
	return &simpleMigration{up, down}
}

func NewStrings(up, down string) Interface {
	return &simpleMigration{String(up), String(down)}
}

// a raw sql query that can be used for up or down.
type String string

func (m String) Exec(db *sql.Tx) error {
	_, err := db.Exec(string(m))
	return err
}

type namedMigration struct {
	Name string
	M    Interface
}
