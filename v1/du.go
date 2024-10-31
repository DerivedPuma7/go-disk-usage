package v1

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path"
)

func init() {
	flag.Parse()
	roots := flag.Args()
	if len(roots) == 0 {
		roots = []string {"."}
	}

	filesizes := make(chan int64)
	go func() {
		for _, root := range roots {
			walkDir(root, filesizes)
		}
		close(filesizes)
	}()

	var nFiles, nBytes int64
	for size := range filesizes {
		nFiles++
		nBytes += size
	}

	printDiskUsage(nFiles, nBytes)
}

func walkDir(dir string, fileSizes chan<- int64) {
	for _, entry := range dirents(dir) {
		if entry.IsDir() {
			subDir := path.Join(dir, entry.Name())
			walkDir(subDir, fileSizes)
		} else {
			fileInfo, _ := entry.Info()
			fileSizes <- fileInfo.Size()
		}
	}
}

func dirents(dir string) []fs.DirEntry {
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "du1: %v\n", err)
		return nil
	}
	return entries
}

func printDiskUsage(nFiles, nBytes int64) {
	fmt.Printf("%d files %.1f GB \n", nFiles, float64(nBytes/1e9))
}
