package v3

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var vFlag = flag.Bool("v", false, "show verbose progress messages")

func init() {
	flag.Parse()
	roots := flag.Args()
	if len(roots) == 0 {
		roots = []string {"."}
	}
	
	filesizes := make(chan int64)
	var n sync.WaitGroup
	for _, root := range roots {
		n.Add(1)
		go walkDir(root, &n, filesizes)
	}
	go func() {
		n.Wait()
		close(filesizes)
	}()

	var tick <-chan time.Time
	if *vFlag {
		tick = time.Tick(500 * time.Millisecond)
	}

	var nFiles, nBytes int64
loop:
	for {
		select {
		case size, ok := <-filesizes:
			if !ok {
				break loop; // break exits both for loop and loop label
			}
			nFiles++
			nBytes += size
		case <-tick :
			printDiskUsage(nFiles, nBytes)
		}
	}
	printDiskUsage(nFiles, nBytes)
}

func walkDir(dir string, n *sync.WaitGroup, fileSizes chan<- int64) {
	defer n.Done()
	for _, entry := range dirents(dir) {
		if entry.IsDir() {
			n.Add(1)
			subDir := filepath.Join(dir, entry.Name())
			go walkDir(subDir, n, fileSizes)
		} else {
			fileInfo, _ := entry.Info()
			fileSizes <- fileInfo.Size()
		}
	}
}

var sema = make(chan struct{}, 20)
func dirents(dir string) []fs.DirEntry {
	sema <- struct{}{}
	defer func() { <-sema }()

	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "du3: %v\n", err)
		return nil
	}
	return entries
}

func printDiskUsage(nFiles, nBytes int64) {
	fmt.Printf("%d files %.1f GB \n", nFiles, float64(nBytes/1e9))
}
