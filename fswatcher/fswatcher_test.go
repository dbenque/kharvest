package fswatcher

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"testing"

	"github.com/fsnotify/fsnotify"
)

func TestFsWatcherCreateModify(t *testing.T) {
	//Create tmp dir
	dir, err := ioutil.TempDir("", "unittest-TestFsWatcherCreateModify")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir) // clean up

	filename := "toto.txt"
	fullpath := path.Join(dir, filename)
	watcher := StartFileWatcher([]string{fullpath})

	content1 := []byte("temporary file's content")
	content2 := []byte("temporary file's content second content")

	next := make(chan struct{})

	go func() {

		event1 := <-watcher.GetEventChan()
		if event1.Op != fsnotify.Create {
			t.Errorf("A Create event was expected. Received %s", event1.Op)
			return
		}
		if string(event1.Content) != string(content1) {
			t.Errorf("Bad content at create.\nExpected: %s\nGot: %s\n", content1, event1.Content)
			return
		}
		next <- struct{}{}
		event2 := <-watcher.GetEventChan()
		if event2.Op != fsnotify.Write {
			t.Errorf("A Create event was expected. Received %s", event2.Op)
			return
		}
		if string(event2.Content) != string(content2) {
			t.Errorf("Bad content at modify.\nExpected: %s\nGot: %s\n", content2, event2.Content)
			return
		}
		close(next)
	}()

	ioutil.WriteFile(fullpath, content1, os.ModePerm)
	<-next
	ioutil.WriteFile(fullpath, content2, os.ModePerm)
	<-next
}

func TestFsWatcherCreateDeleteCreate(t *testing.T) {
	//Create tmp dir
	dir, err := ioutil.TempDir("", "unittest-TestFsWatcherCreateDeleteCreate")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir) // clean up

	filename := "toto.txt"
	fullpath := path.Join(dir, filename)
	watcher := StartFileWatcher([]string{fullpath})
	next := make(chan struct{})

	content1 := []byte("temporary file's content")
	content3 := []byte("temporary file's content second content")

	go func() {
		event1 := <-watcher.GetEventChan()
		if event1.Op != fsnotify.Create {
			t.Errorf("A Create event was expected. Received %s", event1.Op)
			return
		}
		if string(event1.Content) != string(content1) {
			t.Errorf("Bad content at create.\nExpected: %s\nGot: %s\n", content1, event1.Content)
			return
		}

		next <- struct{}{}
		event2 := <-watcher.GetEventChan()
		if event2.Op != fsnotify.Remove {
			t.Errorf("A Create event was expected. Received %s", event2.Op)
			return
		}
		next <- struct{}{}
		event3 := <-watcher.GetEventChan()
		if event3.Op != fsnotify.Create {
			t.Errorf("A Create event was expected. Received %s", event3.Op)
			return
		}
		if string(event3.Content) != string(content3) {
			t.Errorf("Bad content at create.\nExpected: %s\nGot: %s\n", content3, event3.Content)
			return
		}
		close(next)
	}()

	ioutil.WriteFile(fullpath, content1, os.ModePerm)
	<-next
	os.Remove(fullpath)
	<-next
	ioutil.WriteFile(fullpath, content3, os.ModePerm)
	<-next
}
