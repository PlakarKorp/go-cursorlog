package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sync"

	cursorlog "github.com/PlakarKorp/go-cursorlog"
)

func main() {
	state := flag.String("state", "./.cursorlog.json", "path to cursorlog state file")
	flag.Parse()
	files := flag.Args()
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "usage: cursorlogdemo -state <statefile> <file1> [file2 ...]")
		os.Exit(2)
	}

	clog, err := cursorlog.NewCursorLog(*state)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error initializing cursor log: %v\n", err)
		os.Exit(1)
	}
	defer clog.Close()

	wg := sync.WaitGroup{}
	wg.Add(len(files))
	for _, file := range files {
		go func(f string) {
			defer wg.Done()
			rd, err := clog.Open(f)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error tailing %s: %v\n", file, err)
				return
			}

			scanner := bufio.NewScanner(rd)
			for scanner.Scan() {
				line := scanner.Text()
				fmt.Println(line)
			}
			rd.Close()
		}(file)
	}
	wg.Wait()

}
