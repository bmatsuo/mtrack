// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// model_test.go [created: Sat,  3 Aug 2013]

package model

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

var dblock = make(chan string, 1)

func DBTest(t *testing.T, testfn func()) {
	p := make([]byte, base64.URLEncoding.DecodedLen(16))
	_, err := rand.Read(p)
	if err != nil {
		t.Fatal(err)
	}
	uniqId := base64.URLEncoding.EncodeToString(p)
	path := filepath.Join(
		os.TempDir(),
		"mtrack-"+uniqId+".sqlite")
	defer func() {
		err = os.Remove(path)
		if err != nil {
			t.Fatal(err)
		}
	}()

	dblock <- ""
	defer func() { <-dblock }()

	DBPath = path
	err = DBInit()
	if err != nil {
		t.Fatal(err)
	}

	if testfn != nil {
		testfn()
	}
}

func TestInit(t *testing.T) { DBTest(t, nil) }

func TestErrNoRows(t *testing.T) {
	DBTest(t, func() {
		row := DB.QueryRow(`SELECT UserId FROM Users LIMIT 1`)
		var userid string
		err := row.Scan(&userid)
		if err != sql.ErrNoRows {
			t.Fatal("unexpected error:", err)
		}
		rows, err := DB.Query(`SELECT * FROM Users`)
		if err != nil {
			t.Fatal("unexpected error:", err)
		}
		if rows.Next() {
			t.Fatal("unexpected row")
		}
		err = rows.Err()
		if err != nil {
			t.Fatal("unexpected error:", err)
		}
	})
}
