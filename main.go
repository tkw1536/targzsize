// Package targzsize contains the main logic for the targzsize command
package targzsize

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/pkg/errors"
)

//go:generate gogenlicense -p targzsize -n Notices -d notices.go github.com/tkw1536/targzsize

// TotalToString turns value into an optionally human-readable string.
// When human is false, returns a simple string
func TotalToString(value *big.Int, human bool) string {
	if !human {
		return value.String()
	}

	return humanize.BigBytes(value)
}

const chanBufferSize = 100

// MainFile handles a single file, adding the total to total.
func MainFile(filepath string, total *big.Int, silent bool, human bool) error {
	if !silent {
		log.Printf("Reading %s\n", filepath)
	}
	lines := make(chan StatusLine, chanBufferSize)
	items := make(chan Item, chanBufferSize)

	resultChan := ProcessFile(filepath, items)
	countCtx := AddItems(total, items, lines, silent)
	writerCtx := WriteLines(lines, human)

	<-countCtx.Done()
	<-writerCtx.Done()

	return <-resultChan
}

// StatusLine represents a single status line to be written to the output
type StatusLine struct {
	Path  string
	Total big.Int
}

// WriteLines keeps writing from the lines channel to standard output.
// For every write, the existing line is overwritten.
//
// Returns a context that is cancelled when output writing is done.
func WriteLines(lines <-chan StatusLine, human bool) context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer cancel()

		for item := range lines {
			fmt.Fprintf(os.Stderr, "\033[2K\r%s %q", TotalToString(&item.Total, human), item.Path)
		}
	}()

	return ctx
}

// Item represents an item inside a tar.gz file.
type Item struct {
	Path string
	Size int64
}

// AddItems keeps addding to dest from channel values.
// For each add encountered, adds a new status line.
// When silent is set, does not write status lines.
//
// When finished, closes lines.
// Returns a context that is cancelled when adding is finished.
func AddItems(dest *big.Int, items <-chan Item, lines chan<- StatusLine, silent bool) context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer cancel()
		defer close(lines)

		for item := range items {
			dest.Add(dest, big.NewInt(item.Size))
			if silent {
				continue
			}
			lines <- StatusLine{
				Path:  item.Path,
				Total: *dest,
			}
		}
	}()

	return ctx
}

// ProcessFile processes file, writing the size of each chunk containined in it to values.
// Furthermore writes a log message to logChan.
//
// When an error occcurs, calls log.Fattalf.
//
// Returns a channel that receives the error from this function
func ProcessFile(filepath string, items chan<- Item) <-chan error {
	errChan := make(chan error, 1)

	go func() {
		defer close(errChan)
		defer close(items)

		// Open the file
		file, err := os.Open(filepath)
		if err != nil {
			errChan <- errors.Wrapf(err, "Unable to open %s", filepath)
			return
		}
		defer file.Close()

		// make a gzip reader
		gzf, err := gzip.NewReader(file)
		if err != nil {
			errChan <- errors.Wrapf(err, "Unable to create gzip reader")
			return
		}

		// make a tar reader
		tgz := tar.NewReader(gzf)
		if tgz == nil {
			errChan <- errors.New("Unable to create tar reader")
			return
		}

		// iterate over the file
		for {
			header, err := tgz.Next()
			if err == io.EOF {
				break
			}

			if err != nil {
				errChan <- errors.Wrap(err, "Error scanning tarfile")
				return
			}

			switch header.Typeflag {
			case tar.TypeReg:
				items <- Item{
					Size: header.Size,
					Path: header.Name,
				}
			default:
				items <- Item{
					Size: 0,
					Path: header.Name,
				}
			}
		}

	}()
	return errChan
}
