package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const zipExt = ".zip"

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [directory]\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	var dir string
	if len(args) < 1 {
		var err error
		dir, err = os.Getwd()
		panicOnErr(err)
	} else {
		dir = args[0]
	}

	fmt.Printf("Zipping each subdirectory in %s...\n", dir)

	entries, err := ioutil.ReadDir(dir)
	panicOnErr(err)

	wg := sync.WaitGroup{}
	for _, entry := range entries {
		name := entry.Name()

		// Skip if the file is not directory
		if !entry.IsDir() {
			fmt.Printf("%s is not directory; Skipping...\n", name)
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Printf("Zipping %s...\n", name)
			zipDir(filepath.Join(dir, name))
		}()
	}

	wg.Wait()
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func zipDir(dirname string) {
	zipFileName := dirname + zipExt
	fmt.Printf("Zipping %s to %s...\n", dirname, zipFileName)

	zipFile, err := os.Create(zipFileName)
	if err != nil {
		printOnErr(err)
		return
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	err = filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			fmt.Printf("%s is directory; Skipping...\n", path)
			return nil
		}

		// Skip dot files (e.g. .DS_Store)
		if strings.HasPrefix(filepath.Base(path), ".") {
			fmt.Printf("%s is dot file; Skipping...\n", path)
			return nil
		}

		fmt.Println("Adding", path)

		w, err := archive.Create(path)
		if err != nil {
			return fmt.Errorf("error creating writer: %s", err)
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("error opening file: %s", err)
		}
		defer file.Close()

		_, err = io.Copy(w, file)
		return err
	})

	printOnErr(err)
}

func printOnErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
