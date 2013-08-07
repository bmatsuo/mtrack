// Copyright 2013, Bryan Matsuo. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// example_testo.go [created: Wed,  7 Aug 2013]

package migration

import (
	"fmt"
)

// The Sequence type behaves more or less like a slice.
func ExampleSequence() {
	// preallocation with a capacity. like a slice.
	capacity := 10
	seq := MakeSequence(capacity)

	// chained appending. not *that* slice-y.
	seq2 := seq
	seq2 = seq2.Append("20130807-create-examples-table", NewStrings(
		`CREATE TABLE Examples (
			Name VARCHAR[255] PRIMARY KEY
		)`,
		`DROP TABLE Examples`,
	))
	seq2 = seq2.Append("20130809-create-exampledata-table", NewStrings(
		`CREATE TABLE ExampleData (
			Example VARCHAR[255] REFERENCES Examples(Name),
			Datum   INTEGER      NOT NULL
		)`,
		`DROP TABLE ExampleData`,
	))

	// slicing. like a slice
	seq1 := seq2.Slice(1, seq2.Len())

	// slice-like behavior with capacity, appending, and slicing
	fmt.Println(seq.Len(), seq1.Len(), seq2.Len())

	// indexing. like a slice (including the panic).
	name, _ := seq1.Index(0)
	fmt.Println(name)

	// Output:
	// 0 1 2
	// 20130809-create-exampledata-table
}
