package fswatcher

import (
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"sync"

	"io/ioutil"

	"time"

	"github.com/fsnotify/fsnotify"
)

//FileContent notification with the type of modification and content
type FileContent struct {
	Filepath      string
	Content       []byte
	MD5           [md5.Size]byte
	Op            fsnotify.Op
	sendTentative int
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

	return &FileContent{Content: content, MD5: md5.Sum(content), Filepath: filepath}
}

//FileWatcher monitor a given file on the file system and
type FileWatcher struct {
	sync.Mutex
	filepathes map[string]chan struct{}
	resultChan chan *FileContent
	watcher    *fsnotify.Watcher
	stopChan   chan struct{} // Global stop
}

//GetEventChan return the channel for the events related to file changes
func (f *FileWatcher) GetEventChan() <-chan *FileContent {
	return f.resultChan
}

//Resend resend event in one second (max 20 attemps per items)
func (f *FileWatcher) Resend(fc *FileContent) {
	fc.sendTentative++
	if fc.sendTentative < 20 {
		go func() {
			T := time.NewTimer(time.Second)
			defer T.Stop()
			<-T.C
			f.resultChan <- fc
		}()
	} else {
		log.Printf("Error in watcher: more than 20 resend tentative for %s", fc.Filepath)
	}
	return
}

//Stop stop the watcher.
func (f *FileWatcher) Stop() {
	close(f.stopChan)
	f.watcher.Close()
	close(f.resultChan)
}

//StartFileWatcher start a new file watcher for the
func StartFileWatcher(filepathes []string) *FileWatcher {

	mapFiles := map[string]chan struct{}{}

	fileWatcher := &FileWatcher{
		filepathes: mapFiles,
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
				log.Printf("[fswatcher] event: %s", event.String())

				switch event.Op {
				case fsnotify.Create, fsnotify.Write:
					fc := NewFileContent(event.Name)
					if len(fc.Content) == 0 {
						log.Printf("[fswatcher] [warning] empty content")
					}
					fc.Op = event.Op
					fileWatcher.resultChan <- fc
				case fsnotify.Rename, fsnotify.Remove:
					fc := &FileContent{Filepath: event.Name, Content: nil, MD5: [md5.Size]byte{}}
					fc.Op = event.Op
					fileWatcher.resultChan <- fc
					//Restart a filewatch for new create
					fileWatcher.Lock()
					delete(fileWatcher.filepathes, event.Name)
					fileWatcher.Unlock()
					go fileWatcher.startSingleFileWatch(event.Name)
				}
			case err := <-fileWatcher.watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	fileWatcher.Add(filepathes)
	return fileWatcher
}

//Set the list of file to be watched
func (f *FileWatcher) Set(filepaths []string) {
	purgedMap := map[string]chan struct{}{}
	toAdd := []string{}
	f.Lock()
	//Loop and keep intersection in purgedMap
	for _, file := range filepaths {
		if c, ok := f.filepathes[file]; ok {
			purgedMap[file] = c
			delete(f.filepathes, file)
		} else {
			toAdd = append(toAdd, file)
		}
	}

	//everything that remain in the map need to be removed
	for file, cancelChan := range f.filepathes {
		if cancelChan == nil {
			f.watcher.Remove(file)
		} else {
			close(cancelChan)
		}
	}

	//update the map with what remain
	f.filepathes = purgedMap
	f.Unlock()

	//append missing files
	f.Add(toAdd)
}

//Add a new file to watcher
func (f *FileWatcher) Add(filepaths []string) {
	for _, filepath := range filepaths {
		go f.startSingleFileWatch(filepath)
	}
}

//Remove a new file to watcher
func (f *FileWatcher) Remove(filepaths []string) {
	f.Lock()
	defer f.Unlock()
	for _, filepath := range filepaths {
		if cancelChan, ok := f.filepathes[filepath]; !ok {
			fmt.Printf("Unknown: File %s\n", filepath)
			return
		} else if cancelChan == nil {
			f.watcher.Remove(filepath)
		} else {
			close(cancelChan)
		}
		delete(f.filepathes, filepath)
	}
}

//List all the files being watched
func (f *FileWatcher) List() []string {
	f.Lock()
	defer f.Unlock()
	result := []string{}
	for k := range f.filepathes {
		result = append(result, k)
	}
	return result
}
func (f *FileWatcher) startSingleFileWatch(filepath string) {
	f.Lock()
	if _, ok := f.filepathes[filepath]; ok {
		f.Unlock()
		fmt.Printf("Skip: File %s was already registered for watch\n", filepath)
		return
	}
	cancelChan := make(chan struct{})
	f.filepathes[filepath] = cancelChan
	f.Unlock()
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
			f.Lock()
			if cchan, ok := f.filepathes[filepath]; ok && cchan != nil {
				delete(f.filepathes, filepath)
				close(cchan)
			}
			f.Unlock()
		case <-f.stopChan:
			return // this is full stop of the watcher
		case <-cancelChan:
			return // this is because a given file has been removed from the map or a fsnotifier is now running
		}
	}
}
