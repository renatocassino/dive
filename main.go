package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type FileLine struct {
	File string
	Line int
}

var listOfFiles []FileLine

func findWordInBuffer(pattern, path string, scanner *bufio.Scanner) {
	scanner.Split(bufio.ScanLines)

	numberLine := 1
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, pattern) {
			listOfFiles = append(listOfFiles, FileLine{
				Line: numberLine + 1,
				File: path,
			})
		}

		numberLine++
	}
}

func findWordInFile(pattern, path string) error {
	inFile, err := os.Open(path)

	if err != nil {
		return err
	}
	defer inFile.Close()

	scanner := bufio.NewScanner(inFile)

	findWordInBuffer(pattern, path, scanner)

	return nil
}

func printFile(include, pattern string, excludeDir []string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Print("Walking error: ", err)
			return nil
		}

		if info.IsDir() {
			dir := filepath.Base(path)
			for _, d := range excludeDir {
				if d == dir {
					return filepath.SkipDir
				}
			}
		}
		if !info.IsDir() {
			matched, err := filepath.Match(include, info.Name())
			if err != nil {
				fmt.Println("File path matching error: ", err)
				return err
			}
			if matched {
				err = findWordInFile(pattern, path)
				if err != nil {
					log.Print("Error finding word in file: ", err)
				}
			}
		}
		return nil
	}
}

func worker(id int, pattern string, dirsChan chan string, wg *sync.WaitGroup) {
	for {
		select {
		case dir := <-dirsChan:
			files, err := ioutil.ReadDir(dir)
			if err != nil {
				log.Fatal("Reading directory error: ", err)
			}

			for _, file := range files {
				if file.IsDir() {
					go func(dirsC chan string, path string) {
						dirsC <- path
					}(dirsChan, fmt.Sprintf("%s/%s", dir, file.Name()))
					wg.Add(1)
				} else {
					findWordInFile(pattern, fmt.Sprintf("%s/%s", dir, file.Name()))
				}
			}
			wg.Done()
		default:
		}
	}
}

func WalkParrallel(dir, pattern string) {
	var wg sync.WaitGroup

	numWorkers := 4
	if n := runtime.NumCPU(); n > numWorkers {
		numWorkers = n
	}

	dirsChan := make(chan string, numWorkers)

	for w := 1; w <= numWorkers; w++ {
		go worker(w, pattern, dirsChan, &wg)
	}

	go func(dirsC chan string, path string) {
		dirsC <- path
	}(dirsChan, dir)
	wg.Add(1)

	wg.Wait()
}

func main() {
	WalkParrallel(".", "// @brk")

	if len(os.Args) < 2 {
		fmt.Println("You must pass go filename as a parameter.")
		return
	}

	filename := os.Args[1]
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Println(fmt.Sprintf("The filename %s does not exist in this path", filename))
		return
	}

	var content []string
	for index, fileLine := range listOfFiles {
		lineContent := fmt.Sprintf("break name%d %s:%d", index, fileLine.File, fileLine.Line)
		content = append(content, lineContent)
	}

	content = append(content, "continue")
	contentFile := strings.Join(content, "\n")

	err := ioutil.WriteFile("/tmp/__dive-lines", []byte(contentFile), 0777)
	if err != nil {
		panic(err)
	}

	command := fmt.Sprintf("dlv debug %s --init \"/tmp/__dive-lines\"", filename)
	fmt.Println(fmt.Sprintf("Running command: %s", command))
	cmd := exec.Command(command)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err = cmd.Run()

	fmt.Println(err)
}
