package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

var watcher *fsnotify.Watcher
var pattern string
var cmd string

func main() {

	if len(os.Args) < 3 {
		fmt.Printf("usage: go-watch [path] [options] [cli command]\n")
		os.Exit(1)
	}

	p := os.Args[1]

	switch p {
	case "-h", "help", "--help", "--h":
		fmt.Println("\n\n $ go-watch [path] [options] [cli command]")
		fmt.Println()
		fmt.Println()
		fmt.Println(" options: \n\n --ignore=\toptional ignore path pattern, example: ")
		fmt.Println()
		fmt.Println()
		os.Exit(0)
	}

	fmt.Println("Watching: " + p)

	if strings.HasPrefix(os.Args[2], "--ignore=") == true {
		pattern = strings.SplitAfter(os.Args[2], "--ignore=")[1]
		fmt.Println("Ignoring Pattern: " + pattern)
		cmd = strings.Join(os.Args[3:], " ")
	} else {
		pattern = `^\.`
		cmd = strings.Join(os.Args[2:], " ")
	}

	if len(os.Args) >= 3 {
		fmt.Println("Command: " + cmd)
	}

	fmt.Println()

	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Watch Error!")
	}
	defer watcher.Close()

	if err := filepath.Walk(p, walkFunc); err != nil {
		fmt.Println("ERROR", err)
	}

	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Printf("Path Name: %#v, Action Type: %#v \n", event.Name, event.Op.String())
				wg := new(sync.WaitGroup)
				wg.Add(1)
				go exeCmd(cmd, wg)
				wg.Wait()

			case err := <-watcher.Errors:
				fmt.Println("ERROR", err)
			}
		}
	}()

	<-done

}

func walkFunc(path string, fi os.FileInfo, err error) error {

	if path != `.` {
		re := regexp.MustCompile(pattern)
		filename := fi.Name()
		matched := re.MatchString(filename)
		if matched == true {
			return nil
		}
	}

	if fi.Mode().IsDir() {
		return watcher.Add(path)
	}

	return nil
}

func exeCmd(cmd string, wg *sync.WaitGroup) {
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:len(parts)]

	out, err := exec.Command(head, parts...).Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println()
	fmt.Printf("%s", out)
	fmt.Println()
	wg.Done()
}
