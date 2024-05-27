package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	tm "github.com/buger/goterm"
	"github.com/fsnotify/fsnotify"
	ignore "github.com/sabhiram/go-gitignore"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == "--help" || os.Args[1] == "-h" {
		fmt.Print(OptionHelp)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = watcher.Add(".")
	if err != nil {
		panic(err)
	}

	gign, err := os.ReadFile(".gitignore")
	ignored := []string{"node_modules"}
	if err == nil {
		ignored = strings.Split(string(gign), "\n")
	}
	ignoreMatcher := ignore.CompileIgnoreLines(ignored...)

	err = filepath.WalkDir(".", func(walkPath string, fi fs.DirEntry, err error) error {
		if err != nil {
			log.Println(err)
			return nil
		}
		if fi.IsDir() {
			// check if dot directory
			if strings.HasPrefix(walkPath, ".") {
				return nil
			}
			// use .gitignore to ignore directories
			if ignoreMatcher.MatchesPath(walkPath) {
				return nil
			}
			if err = watcher.Add(walkPath); err != nil {
				log.Println(err)
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	cmd := strings.Join(os.Args[1:], " ")
	sh, err := NewShell("/bin/bash")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		_, err := io.Copy(os.Stdout, sh.Stdout)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()
	go func() {
		_, err := io.Copy(os.Stderr, sh.Stderr)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()
	wg := sync.WaitGroup{}

	wg.Add(1)

	tm.Clear()
	tm.MoveCursor(1, 1)
	tm.Println(tm.Color(tm.Bold("** Ctrl-C to exit **"), tm.RED))
	tm.Flush()
	_, err = io.Copy(sh.Stdin, strings.NewReader(cmd+"\n"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer sh.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					log.Println(err)
					wg.Done()
					return
				}
				if ignoreMatcher.MatchesPath(event.Name) {
					return
				}
				log.Printf("%+v", event)

				sh.Close()
				sh, err = NewShell("/bin/bash")
				if err != nil {
					log.Println(err)
					wg.Done()
					return
				}
				go func() {
					_, err := io.Copy(os.Stdout, sh.Stdout)
					if err != nil {
						fmt.Fprintln(os.Stderr, err)
					}
				}()
				go func() {
					_, err := io.Copy(os.Stderr, sh.Stderr)
					if err != nil {
						fmt.Fprintln(os.Stderr, err)
					}
				}()

				tm.Flush()
				tm.Clear()
				tm.MoveCursor(1, 1)
				tm.Println(tm.Color(tm.Bold("Trying to run the command"), tm.GREEN))
				_, err = io.Copy(sh.Stdin, strings.NewReader(cmd+"\n"))
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					wg.Done()
				}
				log.Println("error: ", err, ", ok: ", ok)
			}
		}
	}()

	go func() {
		<-sigs
		wg.Done()
	}()

	wg.Wait()
	close(sigs)
	sh.Close()
}
