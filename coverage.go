package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"golang.org/x/exp/slices"
	"golang.org/x/tools/cover"
)

func MergeCoverage(files []string, output string) {
	w, err := os.Create(output)
	if err != nil {
		fmt.Println("create cover file error:", err)
		os.Exit(-1)
	}
	defer w.Close()
	w.WriteString("mode: set\n")

	result := make(map[string]*cover.Profile)
	for _, file := range files {
		collectOneCoverProfileFile(result, file)
	}

	w1 := bufio.NewWriter(w)
	for _, prof := range result {
		for _, block := range prof.Blocks {
			fmt.Fprintf(w1, "%s:%d.%d,%d.%d %d %d\n",
				prof.FileName,
				block.StartLine,
				block.StartCol,
				block.EndLine,
				block.EndCol,
				block.NumStmt,
				block.Count,
			)
		}
		if err := w1.Flush(); err != nil {
			log.Fatal("flush data to cover profile file error:", err)
		}
	}
}

func collectOneCoverProfileFile(result map[string]*cover.Profile, file string) {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal("open temp cover file error:", err)

	}
	defer f.Close()

	profs, err := cover.ParseProfilesFromReader(f)
	if err != nil {
		log.Fatal("parse cover profile file error:", err)
	}
	mergeProfile(result, profs)
}

func compareProfileBlock(x, y cover.ProfileBlock) int {
	if x.StartLine < y.StartLine {
		return -1
	}
	if x.StartLine > y.StartLine {
		return 1
	}

	// Now x.StartLine == y.StartLine
	if x.StartCol < y.StartCol {
		return -1
	}
	if x.StartCol > y.StartCol {
		return 1
	}

	return 0
}

func mergeProfile(m map[string]*cover.Profile, profs []*cover.Profile) {
	for _, prof := range profs {
		slices.SortFunc(prof.Blocks, func(bi, bj cover.ProfileBlock) bool {
			return bi.StartLine < bj.StartLine || bi.StartLine == bj.StartLine && bi.StartCol < bj.StartCol
		})
		old, ok := m[prof.FileName]
		if !ok {
			m[prof.FileName] = prof
			continue
		}

		// Merge samples from the same location.
		// The data has already been sorted.
		tmp := old.Blocks[:0]
		var i, j int
		for i < len(old.Blocks) && j < len(prof.Blocks) {
			v1 := old.Blocks[i]
			v2 := prof.Blocks[j]

			switch compareProfileBlock(v1, v2) {
			case -1:
				tmp = appendWithReduce(tmp, v1)
				i++
			case 1:
				tmp = appendWithReduce(tmp, v2)
				j++
			default:
				tmp = appendWithReduce(tmp, v1)
				tmp = appendWithReduce(tmp, v2)
				i++
				j++
			}
		}
		for ; i < len(old.Blocks); i++ {
			tmp = appendWithReduce(tmp, old.Blocks[i])
		}
		for ; j < len(prof.Blocks); j++ {
			tmp = appendWithReduce(tmp, prof.Blocks[j])
		}

		m[prof.FileName] = old
	}
}

// appendWithReduce works like append(), but it merge the duplicated values.
func appendWithReduce(input []cover.ProfileBlock, b cover.ProfileBlock) []cover.ProfileBlock {
	if len(input) >= 1 {
		last := &input[len(input)-1]
		if b.StartLine == last.StartLine &&
			b.StartCol == last.StartCol &&
			b.EndLine == last.EndLine &&
			b.EndCol == last.EndCol {
			if b.NumStmt != last.NumStmt {
				panic(fmt.Errorf("inconsistent NumStmt: changed from %d to %d", last.NumStmt, b.NumStmt))
			}
			// Merge the data with the last one of the slice.
			last.Count |= b.Count
			return input
		}
	}
	return append(input, b)
}
