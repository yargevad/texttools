package main

import (
	"encoding/csv"
	"flag"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

/* From `cut` docs:
The [fieldSpec] option argument is a comma or whitespace separated set of
numbers and/or number ranges. Number ranges consist of a number, a dash (`-'),
and a second number and select the fields or columns from the first number to
the second, inclusive. Numbers or number ranges may be preceded by a dash, which
selects all fields or columns from 1 to the last number. Numbers and number
ranges may be repeated, overlapping, and in any order.

TODO: Numbers or number ranges may be followed by a dash, which selects all fields or columns from the last number to the end of the line. (note: this is why int64 is used instead of int, so the `finish` portion can be math.MaxInt64)
TODO: If a field or column is specified multiple times, it will appear only once in the output.
TODO: It is not an error to select fields or columns not present in the input line.
*/

func main() {
	var colSep *regexp.Regexp = regexp.MustCompile(`\s+|\s*,\s*`)
	var in = flag.String("in", "", "csv filename to read")
	var fieldSpec = flag.String("f", "", "which fields to output")

	flag.Parse()
	args := flag.Args()
	// assume first argument after flags is filename
	if len(args) >= 1 {
		in = &args[0]
	}

	if in == nil || strings.TrimSpace(*in) == "" {
		log.Fatal("FATAL: --in is a required parameter\n")
	}

	var cols []int64
	if fieldSpec != nil && strings.TrimSpace(*fieldSpec) != "" {
		colSpecs := colSep.Split(*fieldSpec, -1)
		for _, colSpec := range colSpecs {
			if strings.Contains(colSpec, "-") {
				// we have a range
				var start, finish int64
				r := strings.Split(colSpec, "-")
				if len(r) > 2 {
					log.Fatal("FATAL: unexpected range format [%s]\n")
				}

				if strings.TrimSpace(r[0]) == "" {
					start = int64(0)
				} else {
					c, err := strconv.Atoi(r[0])
					if err != nil {
						log.Fatal(err)
					}
					start = int64(c)
				}

				if strings.TrimSpace(r[1]) == "" {
					log.Fatal("FATAL: empty ending ranges are not supported\n")
				} else {
					c, err := strconv.Atoi(r[1])
					if err != nil {
						log.Fatal(err)
					}
					finish = int64(c)
					tmpCols := make([]int64, 0, finish-start+1)
					for i := start; i <= finish; i++ {
						tmpCols = append(tmpCols, i-1)
					}
					cols = append(cols, tmpCols...)
				}

			} else {
				// single column number
				c, err := strconv.Atoi(colSpec)
				if err != nil {
					log.Fatal(err)
				}
				cols = append(cols, int64(c-1))
			}
		}
	}

	fh, err := os.Open(*in)
	if err != nil {
		log.Fatal(err)
	}

	csvIn := csv.NewReader(fh)
	csvOut := csv.NewWriter(os.Stdout)

	csvData := make([]string, len(cols))
	for {
		// encoding/csv asserts that each line has the same number of columns
		csvRead, err := csvIn.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		for i, idx := range cols {
			csvData[i] = csvRead[idx]
		}

		err = csvOut.Write(csvData)
		if err != nil {
			log.Fatal(err)
		}
	}
	csvOut.Flush()
}
