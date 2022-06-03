package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
)

type ResultFile struct {
	Coverage []string
	Junit    []string
}

func main() {
	ex, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	root_path, err := filepath.EvalSymlinks(path.Join(ex, "bazel-testlogs"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("root: %v\n", root_path)
	var resultFiles ResultFile
	err = filepath.WalkDir(root_path, func(path string, d fs.DirEntry, err error) error {
		if d == nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() == "test.xml" {
			resultFiles.Junit = append(resultFiles.Junit, path)
		}
		if d.Name() == "coverage.dat" {
			resultFiles.Coverage = append(resultFiles.Coverage, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
		return
	}
	MergeCoverage(resultFiles.Coverage, "coverage.dat")
	MergeJunit(resultFiles.Junit, "bazel.xml")
	fmt.Println("complete to collect bazel result.")
}
