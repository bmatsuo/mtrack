// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// scan.go [created: Sat, 27 Jul 2013]

package main

import (
	"crypto/sha1"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/bmatsuo/mtrack/model"
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
		mediaid, err := model.SyncMedia(path, info)
		if err != nil {
			log.Printf("%q (%v): %T %v", path, mediaid, err, err)
		}
		return err
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
