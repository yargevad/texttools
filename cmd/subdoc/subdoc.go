// subdoc takes a list of filenames, compares their content,
// and reports cases where the entire content of one file
// is also present in another file.
//
// file content can optionally be treated as json objects with --json-key;
// in that case, the content of the specified json key is compared, instead
// of the entire content of each file.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"

	"github.com/buger/jsonparser"
	"github.com/fatih/color"
)

type File struct {
	Filename string
	Contents []byte
}
type Files []File

func (f Files) Len() int {
	return len(Files(f))
}

func (f Files) Less(i, j int) bool {
	return len(f[i].Contents) < len(f[j].Contents)
}

func (f Files) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

var jsonKey = flag.String("json-key", "", "json key to look up in files (optional)")

func main() {
	flag.Parse()
	files := []File{}

	// load in the contents of all files specified on the command line
	args := flag.Args()
	if len(args) <= 1 {
		return
	}
	for _, arg := range args[1:] {
		file := File{Filename: arg}
		contents, err := ioutil.ReadFile(arg)
		if err != nil {
			log.Fatal(err)
		}
		file.Contents = contents
		files = append(files, file)
	}

	sort.Sort(sort.Reverse(Files(files)))
	for idx1, f1 := range files {
		if *jsonKey != "" {
			fmt.Fprintf(os.Stdout, "% 8d %s (%s)\n", len(f1.Contents), f1.Filename, *jsonKey)
		} else {
			fmt.Fprintf(os.Stdout, "% 8d %s\n", len(f1.Contents), f1.Filename)
		}
		if idx1 < len(files)-1 {
			var val1 []byte
			var err error
			if *jsonKey != "" {
				val1, _, _, err = jsonparser.Get(f1.Contents, *jsonKey)
				if err != nil {
					log.Fatalf("error getting '%s' from [%s]: %s\n", *jsonKey, f1.Filename, err)
				}
			} else {
				val1 = f1.Contents
			}

			for _, f2 := range files[idx1+1:] {
				var val2 []byte
				var err error
				if *jsonKey != "" {
					val2, _, _, err = jsonparser.Get(f2.Contents, *jsonKey)
					if err != nil {
						log.Fatal("error getting body from [%s]: %s\n", f2.Filename, err)
					}
				} else {
					val2 = f2.Contents
				}

				fidx := bytes.Index(val1, val2)
				if fidx == -1 {
					continue
				}

				fmt.Fprintf(os.Stdout, "           %s\n", color.RedString(f2.Filename))
			}
		}
	}
}
