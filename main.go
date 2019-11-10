package main

import (
	"crypto/sha512"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type duplicateGroupInfo struct {
	size int64
	list []string
}

const version = "1.0.1"

var sameSizeFileList []string // files that have at least one file with same length
var fileSizeBucket = make(map[int64][]string)
var fileHashesPathMap = make(map[[sha512.Size]byte]*duplicateGroupInfo)
var totalFileCount int
var totalDuplicateCount int
var totalDuplicateGroupCount int
var exts []string

func checkDuplicate(pos int, path string) error {
	hasher := sha512.New()
	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer f.Close()
	if _, err := io.Copy(hasher, f); err != nil {
		log.Fatal(err)
	}

	var hash [sha512.Size]byte
	copy(hash[:], hasher.Sum(nil))

	if v, ok := fileHashesPathMap[hash]; ok {
		fmt.Printf("[%d / %d] %s is a duplicate of %s\n", pos, len(sameSizeFileList), path, v.list[0])
		v.list = append(v.list, path)
		if len(v.list) == 2 {
			totalDuplicateCount += 2
			totalDuplicateGroupCount++
		} else {
			totalDuplicateCount++
		}
	} else {
		info, err := f.Stat()
		if err != nil {
			fmt.Println(err)
			return nil
		}
		fileHashesPathMap[hash] = &duplicateGroupInfo{info.Size(), []string{path}}
	}

	return nil
}

func checkFileLength(path string, info os.FileInfo, err error) error {
	if err != nil {
		fmt.Println(err)
		return nil
	}
	if info.IsDir() { // skip directory
		return nil
	}
	if len(exts) > 0 {
		matchSuffix := false
		for _, ext := range exts {
			if strings.HasSuffix(strings.ToLower(path), ext) {
				matchSuffix = true
				break
			}
		}

		if !matchSuffix {
			return nil
		}
	}

	fileSizeBucket[info.Size()] = append(fileSizeBucket[info.Size()], path)
	totalFileCount++
	return nil
}

func deleteDuplicate() {
	for hash, info := range fileHashesPathMap {
		if len(info.list) < 2 {
			continue
		}

		fmt.Printf("\nGroup %s: %d copies (%d MB each)\n", hex.EncodeToString(hash[:])[:6], len(info.list), info.size/1024/1024)
		for _, path := range info.list[1:] {
			fmt.Printf("  Deleting copy %s\n", path)
			err := os.Remove(path)

			if err != nil {
				fmt.Println(err)
			}
		}
	}

	os.Stdout.Sync()
}

func printUsage() {
	fmt.Println("redundantFinder [-r] [-ext=extensions] <target_directory> ...")
	fmt.Println(strings.Replace("Copyright(C) 2019 Ceeji Cheng <hi.ceeji#gmail.com> and contributors", "#", "@", 1))
	fmt.Println()

	flag.PrintDefaults()
}

func parseCLI() (dirs []string, ext []string, delete bool) {
	r := flag.Bool("r", false, "delete redundant copies after scan")
	v := flag.Bool("v", false, "show version and exit")
	exts := flag.String("ext", "jpg|png|arw|raw|nec|jpeg|mp4|mp3|json|m4a|avi|mpeg|mpg|dat|doc|docx|ppt|pptx|db|txt|zip|gz|bz|7z|tar|rar|bzip|iso|pkg|wav", "specify file extensions for scanning, any file without these extension will be ignored. Multiple values should be splited by '|'. If empty, any file will be included.")
	flag.Parse()

	if *v {
		fmt.Println("version " + version + "_" + runtime.Compiler + "_" + runtime.GOOS + "_" + runtime.GOARCH)
		os.Exit(0)
	}

	dirs = flag.Args()

	if len(dirs) == 0 {
		printUsage()
		os.Exit(-1)
	}

	return dirs, strings.Split(*exts, "|"), *r
}

func main() {
	// parse command line
	dirs, _exts, delete := parseCLI()
	exts = _exts

	fmt.Print("Step 1: Scanning Possible Duplicate Files...")

	for _, dir := range dirs {
		err := filepath.Walk(dir, checkFileLength)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	for length, paths := range fileSizeBucket {
		if len(paths) > 1 && length > 0 {
			sameSizeFileList = append(sameSizeFileList, paths...)
		}
	}
	// relax fileSizeBucket
	fileSizeBucket = nil

	fmt.Printf("%d / %d files are possibly duplicate.\n", len(sameSizeFileList), totalFileCount)

	sort.Strings(sameSizeFileList)

	fmt.Println("Step 2: Checking file content...")
	for i, path := range sameSizeFileList {
		err := checkDuplicate(i+1, path)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	fmt.Printf("Finish, %d group files has %d copies, %d will be deleted.\n", totalDuplicateGroupCount, totalDuplicateCount, totalDuplicateCount-totalDuplicateGroupCount)

	// delete files
	if delete {
		deleteDuplicate()
	} else {
		fmt.Println()
		fmt.Println("Add -r option to remove redundant files.")
	}
}
