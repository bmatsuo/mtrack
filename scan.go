// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// scan.go [created: Sat, 27 Jul 2013]

package main

import (
	"crypto/sha1"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var MediaRoots []FSRoot

type FSRoot struct {
	Name string
	Path string
}

func ScanMedia() {
	for _, root := range MediaRoots {
		log.Printf("Scanning %q", root.Name)
		err := scanMedia(root.Path)
		if err != nil {
			log.Printf("%s: %v", root.Name, err)
		}
	}
	log.Print("Scan complete")
}

func scanMedia(root string) error {
	mediahandler := func(path string, info os.FileInfo) error {
		pathnorm := strings.ToLower(path)
		sha1 := getsha1(pathnorm)
		var mod time.Time
		row := DB.QueryRow(`select ModTime from Media where MediaId = ?`, sha1)
		err := row.Scan(&mod)
		switch err {
		case sql.ErrNoRows:
			q := `INSERT INTO Media(MediaId, Path, PathNorm, ModTime)`
			q += ` VALUES (?, ?, ?, ?)`
			_, err = DB.Exec(q,
				sha1, path, pathnorm, info.ModTime())
			if err != nil {
				log.Printf("%q (%v): DB insert error %T: %v",
					path, sha1, err, err)
			}
		case nil:
			_mod := info.ModTime()
			if _mod.After(mod) {
				q := `UPDATE Media SET ModTime = ? WHERE MediaId = ?`
				_, err = DB.Exec(q, _mod, sha1)
				if err != nil {
					log.Printf("%q (%v): DB update error %T: %v",
						path, sha1, err, err)
				}
			}
		default:
			log.Printf("%q (%v): DB lookup error %T: %v",
				path, sha1, err, err)
			return nil
		}
		return (err)
	}
	return WalkDir(root, []string{".go"}, mediahandler)
}

func getsha1(path string) string {
	h := sha1.New()
	h.Write([]byte(path))
	sum := h.Sum(nil)
	return fmt.Sprintf("%x", sum)
}

func WalkDir(dir string, ext []string, fn func(path string, info os.FileInfo) error) error {
	return filepath.Walk(dir, makeWalker(ext, fn))
}

func makeWalker(ext []string, fn func(string, os.FileInfo) error) filepath.WalkFunc {
	acceptExt := make(map[string]bool, len(ext))
	for _, ext := range ext {
		acceptExt[ext] = true
	}

	return func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if acceptExt[filepath.Ext(path)] {
			fn(path, info)
		}
		return nil
	}
}
