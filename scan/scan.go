// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// scan.go [created: Sat, 27 Jul 2013]

package scan

import (
	"crypto/sha1"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/bmatsuo/mtrack/model"
)

var MediaRoots []FSRoot

type FSRoot struct {
	Name string
	Path string
}

type Root struct {
	Name string   `json:"name"` // for logging purposes
	Path string   `json:"path"`
	Exts []string `json:"exts"`
	Scan struct {
		Delay uint64
	}
}

// Scans the filesystem looking for media files
type Scanner struct {
	roots []*Root
}

// Create a new scanner that searches directories in roots.
func NewScanner(roots []*Root) *Scanner {
	// this feels weird
	return &Scanner{roots}
}

// Recursively scan root directories. Symbolic links are not followed.
func (s *Scanner) Scan() {
	errch := make(chan error, 0)

	for _, root := range s.roots {
		log.Printf("Scan: %q", root.Name)
		go scanMedia(errch, root.Path, root.Exts...)
	}

	for err := range errch {
		if err == nil {
			break
		}
		log.Printf("Scan: %v", err)
	}

	log.Print("Scan: complete")
}

// The error returned when an attempt is call Close() a closed *Cron.
var ErrClosed = fmt.Errorf("closed")

var DefaultScanner *Scanner
var DefaultCron *Cron

func defaultCron() *Cron {
	if DefaultCron == nil {
		panic("nil DefaultCron")
	}
	return DefaultCron
}

func defaultScanner() *Scanner {
	if DefaultScanner == nil {
		panic("nil DefaultScanner")
	}
	return DefaultScanner
}

// Semantically equivalent to
//		DefaultScanner = NewScanner(roots)
//		DefaultCron = NewCron(delay, DefaultScanner)
// Init() must be called before Start().
func Init(delay time.Duration, roots []*Root) {
	DefaultScanner = NewScanner(roots)
	DefaultCron = NewCron(delay, DefaultScanner)
}

// Starts DefaultCron. Panics if DefaultCron is nil.
func Start(statch chan *CronStatus) {
	defaultCron().Start(statch)
}

// Closes DefaultCron. Panics if DefaultCron is nil.
func Close() error {
	return defaultCron().Close()
}

// Scans using DefaultScanner. Panics if DefaultScanner is nil.
func Scan() {
	defaultScanner().Scan()
}

// Scans the filesystem periodically for new media.
type Cron struct {
	scanner  *Scanner
	delay    time.Duration
	term     chan chan error
	termlock chan chan chan error
}

// Statics about a Cron instance.
type CronStatus struct {
	CronStart   time.Time
	LastScan    time.Time
	LastScanDur time.Duration
}

// Create a cron that scans with a delay-minute pause between scans.
// Scanning does not begin until Start() is called.
func NewCron(delay time.Duration, scanner *Scanner) *Cron {
	cron := new(Cron)
	cron.scanner = scanner
	cron.delay = delay
	cron.term = make(chan chan error, 0) // cannot be buffered
	cron.termlock = make(chan chan chan error, 1)
	cron.termlock <- cron.term
	return cron
}

// TODO
func (cron *Cron) Status() *CronStatus {
	return new(CronStatus)
}

// Begin periodically scanning the filesystem for new files.
// This method is not capable of detecting deleted files.
func (cron *Cron) Start(statch chan *CronStatus) {
	go func() { cron.start(statch) }()
}

func (cron *Cron) start(statch chan *CronStatus) {
	start := time.Now()
	term := cron.term
	var timer <-chan time.Time
	var pending *CronStatus
	var backlog []*CronStatus
	var _statch chan *CronStatus
	var termresp []chan error

	// prime the timer to start scanning immediately
	_timer := make(chan time.Time, 1)
	_timer <- time.Now()
	timer = _timer
	for {
		select {
		case _statch <- pending:
			if len(backlog) > 0 {
				pending = backlog[0]
				backlog = backlog[1:]
			} else {
				pending = nil
				_statch = nil
			}
		case <-timer:
			scanstart := time.Now()
			cron.scanner.Scan()
			timer = time.After(cron.delay) // start immediately
			scandur := time.Since(scanstart)
			if statch != nil {
				status := new(CronStatus)
				status.CronStart = start
				status.LastScan = scanstart
				status.LastScanDur = scandur
				if pending == nil {
					pending = status
					_statch = statch
				} else {
					backlog = append(backlog, status)
				}
			}
		case errch := <-term:
			termresp = append(termresp, errch)

			if cron.term != nil {
				<-cron.termlock
				close(cron.termlock)
				close(cron.term)
				cron.term = nil
			}

			errch <- nil

			return // ok because cron.term is unbuffered
		}
	}
}

// this will block until the Cron has closed. if cron.Start() has not previously
// been called the behavior of Close() is not defined; a client should be
// prepared for either an error or blocking until cron.Start() is called.
func (cron *Cron) Close() error {
	term, ok := <-cron.termlock
	if !ok {
		return ErrClosed
	}

	defer func() { cron.termlock <- term }()

	errch := make(chan error, 1)
	term <- errch
	return <-errch
}

// Perform a single scan of the filesystem for media content.
func ScanMedia() {
	errch := make(chan error, 0)

	for _, root := range MediaRoots {
		log.Printf("Scan: %q", root.Name)
		go scanMedia(errch, root.Path, ".go")
	}

	for err := range errch {
		if err == nil {
			break
		}
		log.Printf("Scan: %v", err)
		errch <- err
	}

	log.Print("Scan: complete")
}

func scanMedia(errch chan error, root string, exts ...string) {
	mediahandler := func(path string, info os.FileInfo) error {
		mediaid, err := model.SyncMedia(root, path, info)
		if err != nil {
			errch <- fmt.Errorf("%q (%v): %T %v", path, mediaid, err, err)
		}
		return nil
	}
	err := WalkDir(root, exts, mediahandler)
	errch <- err
	if err != nil {
		// chan still needs "zeroing out"
		errch <- nil
	}
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
