package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/krzhapps/GithubTickets/internal/cli"
)

func main() {
	err := cli.Execute()
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "tickets: %v\n", err)
	var ex *cli.ExitError
	if errors.As(err, &ex) {
		os.Exit(ex.Code)
	}
	os.Exit(1)
}
