package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

type ResultFile struct {
	Coverage []string
	Junit    []string
}

func main() {
	bazelSymlinkPrefix := flag.String("bazel-symlink-prefix", "bazel-", "Bazel symlink prefix.") // For --symlink_prefix in bazel
	coverageFile := flag.String("output-coverage", "coverage.dat", "Output coverage file")
	junitFile := flag.String("output-junit", "bazel.xml", "Output junit file")
	flag.Parse()

	workspaceDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	testLogsDir := fmt.Sprintf("%stestlogs", *bazelSymlinkPrefix)
	rootPath, err := filepath.EvalSymlinks(filepath.Join(workspaceDir, testLogsDir))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("root: %v\n", rootPath)

	var resultFiles ResultFile
	err = filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
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
	}

	MergeCoverage(resultFiles.Coverage, *coverageFile)
	MergeJunit(resultFiles.Junit, *junitFile)
	fmt.Println("complete to collect bazel result.")
}
