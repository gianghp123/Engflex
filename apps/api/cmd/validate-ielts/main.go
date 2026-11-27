// validate-ielts validates staged exam content without writing to the database.
package main

import (
	"fmt"
	"io"
	"os"

	"engflex-api/internal/exams/ielts"
)

func main() {
	raw, err := io.ReadAll(io.LimitReader(os.Stdin, 32<<20))
	if err == nil {
		_, err = ielts.ParseContent(raw)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
