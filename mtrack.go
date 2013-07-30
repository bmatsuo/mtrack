// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// mtrack.go [created: Sat, 27 Jul 2013]

package main

import (
	"log"

	"github.com/bmatsuo/mtrack/config"
	"github.com/bmatsuo/mtrack/http"
	"github.com/bmatsuo/mtrack/model"
	"github.com/bmatsuo/mtrack/scan"
)

func Check(err error) {
	if err != nil {
		log.Fatalf("Fatal error: %v", err)
	}
}

func CheckType(err error) {
	if err != nil {
		log.Fatalf("Fatal error: %T %v", err, err)
	}
}

func main() {
	Check(config.Configure())
	Check(model.DBInit())
	go scan.ScanMedia()
	Check(http.HTTPStart())
}
