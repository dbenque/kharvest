package fswatcher

import (
	"crypto/md5"
	"fmt"
	"log"
	"os"

	"io/ioutil"

	"time"

	"github.com/fsnotify/fsnotify"
)

//FileContent notification with the type of modification and content
type FileContent struct {
	Content []byte
	MD5     [md5.Size]byte
	Op      fsnotify.Op
}

//NewFileContent builds a new filecontent for the given paths
func NewFileContent(filepath string) *FileContent {

	f, err := os.Open(filepath)
	if err != nil {
		log.Printf("error opening file %s: %v", filepath, err)
		return nil
	}
	defer f.Close()

	content, err := ioutil.ReadAll(f)
	if err != nil {
		log.Printf("error reading file %s: %v", filepath, err)
		return nil
	}

	return &FileContent{Content: content, MD5: md5.Sum(content)}
}

//FileWatcher monitor a given file on the file system and
type FileWatcher struct {
	filepathes []string
	resultChan chan *FileContent
	watcher    *fsnotify.Watcher
	stopChan   chan struct{}
}

//GetEventChan return the channel for the events related to file changes
func (f *FileWatcher) GetEventChan() <-chan *FileContent {
	return f.resultChan
}

//Stop stop the watcher.
func (f *FileWatcher) Stop() {
	close(f.stopChan)
	f.watcher.Close()
}

//StartFileWatcher start a new file watcher for the
func StartFileWatcher(filepathes []string) *FileWatcher {

	fileWatcher := &FileWatcher{
		filepathes: filepathes,
		resultChan: make(chan *FileContent, 10*len(filepathes)),
		stopChan:   make(chan struct{}),
	}

	var err error
	fileWatcher.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Printf("error in watcher: %v", err)
	}

	go func() {
		for {
			select {
			case event, ok := <-fileWatcher.watcher.Events:
				if !ok {
					return
				}
				log.Printf("event: %#v", event)
				switch event.Op {
				case fsnotify.Create, fsnotify.Write:
					fc := NewFileContent(event.Name)
					fc.Op = event.Op
					fileWatcher.resultChan <- fc
				case fsnotify.Rename, fsnotify.Remove:
					fc := &FileContent{Content: nil, MD5: [md5.Size]byte{}}
					fc.Op = event.Op
					fileWatcher.resultChan <- fc
					//Restart a filewatch for new create
					go fileWatcher.startSingleFileWatch(event.Name)
				}
			case err := <-fileWatcher.watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	for _, f := range filepathes {
		go fileWatcher.startSingleFileWatch(f)
	}
	return fileWatcher
}

func (f *FileWatcher) startSingleFileWatch(filepath string) {
	tick := time.NewTicker(time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			fmt.Printf("Stat %s\n", filepath)
			if _, err := os.Stat(filepath); err != nil {
				continue
			}
			err := f.watcher.Add(filepath)
			if err != nil {
				log.Printf("error in watcher [%s] : %v", filepath, err)
				continue
			}
			fmt.Printf("Watcher add %s\n", filepath)
			// The create event is never sent by the fsnotifier, let's do it here
			fc := NewFileContent(filepath)
			fc.Op = fsnotify.Create
			f.resultChan <- fc
			return
		case <-f.stopChan:
			return
		}
	}
}
