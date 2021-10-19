// Command targzsize computes the total unpacked size of a set of tar.gz archives.
//
//   targzsize [-legal] [-no-progress] [-human] path [path...]
//
// Targzsize iterates over the provides paths and computes the unpacked size of each file within the packages archives.
// It then adds these totals together and outputs it to standard output.
//
// By default, targzsize writes status messages to standard error.
// Pass the '-no-progress' flag to prevent this.
//
// By default the standard output will contain a single number, representing the total size in bytes.
// To instead use human readable units, pass the '-human' flag.
// This flag also applies to status messages.
//
// The '-legal' flag can be used to print legal and licensing information.
package main

import (
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/tkw1536/targzsize"
)

var silentFlag bool
var humanFlag bool

func main() {
	// get list of files
	files := flag.Args()
	if len(files) == 0 {
		log.Fatal("Need at least one file. ")
	}

	// handle all the files
	var total big.Int
	for _, filepath := range files {
		if err := targzsize.MainFile(filepath, &total, silentFlag, humanFlag); err != nil {
			log.Fatalf("Error processing %s: %s\n", filepath, err)
			return
		}
	}

	// and write the total
	log.Printf("%s\n", targzsize.TotalToString(&total, humanFlag))
}

func init() {
	var legalFlag bool
	flag.BoolVar(&legalFlag, "legal", legalFlag, "Print legal information and exit")
	defer func() {
		if legalFlag {
			fmt.Println("targzsize is licensed under the terms of MIT License.")
			fmt.Println(targzsize.Notices)
			os.Exit(0)
		}
	}()

	defer flag.Parse()

	flag.BoolVar(&silentFlag, "no-progress", silentFlag, "Don't output status messages to stderr")
	flag.BoolVar(&humanFlag, "human", humanFlag, "Output human units instead of bytes")
}
