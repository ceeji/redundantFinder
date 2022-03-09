package main

import (
	"crypto/sha512"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

type duplicateGroupInfo struct {
	size int64
	list []string
}

const (
	version               = "1.0.1"
	hashLengthInSmallMode = 2 * 1024 // 2kb
)

var sameSizeFileList []string // files that have at least one file with same length
var fileSizeBucket = make(map[int64][]string)
var fileHashesPathMap = make(map[[sha512.Size]byte]*duplicateGroupInfo)
var totalFileCount int
var totalDuplicateCount int
var totalDuplicateGroupCount int
var exts []string

type HashMode int8

const (
	HashModeFull  HashMode = 0
	HashModeSmall HashMode = 1
)

func checkDuplicate(pos int, path string, hashMode HashMode) error {
	// initialize file
	hasher := sha512.New()
	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer f.Close()
	var info os.FileInfo

	// read contents and calc hash
	if hashMode == HashModeFull {
		// for full mode, read all contents and generate hash of the file
		_, err = io.Copy(hasher, f)
	} else {
		// for small mode, only read contents in the middle of the file
		if info, err = f.Stat(); err != nil {
			fmt.Println(err)
			return nil
		}
		var hashStart int64
		var size = info.Size()
		if size >= hashLengthInSmallMode*2 {
			hashStart = size / 2
		} else {
			hashStart = size - hashLengthInSmallMode
		}
		f.Seek(hashStart, 0)
		_, err = io.CopyN(hasher, f, hashLengthInSmallMode)
		if err == io.EOF { // ignore End of File
			err = nil
		}
	}

	if err != nil {
		log.Fatal(err)
	}

	var hash [sha512.Size]byte
	copy(hash[:], hasher.Sum(nil))

	if v, ok := fileHashesPathMap[hash]; ok {
		if hashMode == HashModeFull {
			fmt.Printf("[%d / %d] %s is a duplicate of %s\n", pos, len(sameSizeFileList), path, v.list[0])
		}
		v.list = append(v.list, path)
		if len(v.list) == 2 {
			totalDuplicateCount += 2
			totalDuplicateGroupCount++
		} else {
			totalDuplicateCount++
		}
	} else {
		if info == nil {
			info, err = f.Stat()
			if err != nil {
				fmt.Println(err)
				return nil
			}
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
	// ignore hidden and thumbnail files
	if strings.Contains(path, "@__thumb") || strings.Contains(path, "/.") || strings.Contains(path, "@Recently-Snapshot") || strings.Contains(path, "@Recycle") {
		return nil
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

func parseCLI() (dirs []string, ext []string, delete bool, disableSmallHash bool) {
	r := flag.Bool("r", false, "delete redundant copies after scan")
	v := flag.Bool("v", false, "show version and exit")
	p := flag.Bool("disable-smallhash", false, "disable smallhash, which is used to read only a part of the file to quickly exclude most of unique file")
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

	return dirs, strings.Split(*exts, "|"), *r, *p
}

func testBar() {
	count := 100000

	// create and start new bar
	// bar := pb.StartNew(count)

	// start bar from 'default' template
	// bar := pb.Default.Start(count)

	// start bar from 'simple' template
	// bar := pb.Simple.Start(count)

	// start bar from 'full' template
	bar := pb.Full.Start(count)

	for i := 0; i < count; i++ {
		bar.Increment()
		time.Sleep(time.Millisecond)
	}

	// finish bar
	bar.Finish()
}

func main() {
	// parse command line
	dirs, _exts, shouldDelete, disableSmallHash := parseCLI()
	exts = _exts

	// start working
	startTime := time.Now()
	fmt.Print("Step 1: Scanning Possibly Duplicate Files...")

	// testBar()

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

	sort.Strings(sameSizeFileList)

	// using small hash to exclude some files
	if disableSmallHash == false {
		bar := pb.Full.Start(len(sameSizeFileList))
		for i, path := range sameSizeFileList {
			err := checkDuplicate(i+1, path, HashModeSmall)
			if err != nil {
				fmt.Println()
				fmt.Println(err)
			}
			bar.Increment()
		}

		// extract files only for those which have at least one file with duplicate small hash
		sameSizeFileList := []string{}
		for _, info := range fileHashesPathMap {
			if len(info.list) > 1 {
				sameSizeFileList = append(sameSizeFileList, info.list...)
			}
		}
		bar.Finish()
	}

	fmt.Printf("%d / %d files are possibly duplicate.\n", len(sameSizeFileList), totalFileCount)

	fmt.Println("Step 2: Checking file content...")
	bar := pb.Full.Start(len(sameSizeFileList))
	for i, path := range sameSizeFileList {
		err := checkDuplicate(i+1, path, HashModeFull)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		bar.Increment()
	}
	bar.Finish()

	fmt.Printf("Finish, %d group files has %d copies, %d will be deleted, time consuming: %v.\n", totalDuplicateGroupCount, totalDuplicateCount, totalDuplicateCount-totalDuplicateGroupCount, time.Now().Sub(startTime))

	// shouldDelete files
	if shouldDelete {
		deleteDuplicate()
	} else {
		fmt.Println()
		fmt.Println("Add -r option to remove redundant files.")
	}
}
