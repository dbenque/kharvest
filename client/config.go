package client

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/dbenque/kharvest/fswatcher"
)

const (
	ConfigPath = "/etc/kharvest.conf"
)

//ListOfFile list of files given in arguments of the program
type ListOfFile []string

//Set to decode param
func (i *ListOfFile) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func (i *ListOfFile) String() string {
	return fmt.Sprintf("%v", *i)
}

//Config Configuration for file watching
type Config struct {
	Files      ListOfFile
	filesChan  chan []string
	ConfigPath string
	stopChan   chan struct{}
}

//NewConfig create an empty config object
func NewConfig() *Config {
	return &Config{Files: []string{}, ConfigPath: "", filesChan: make(chan []string, 2)}
}

//StopWatching will stop watching configuration file if watch was running
func (c *Config) StopWatching() {
	if c.stopChan != nil {
		close(c.stopChan)
		c.stopChan = nil
	}
}

//ReadAndWatch read the configuration and if it is a config file, watch the configuration to apply changes live.
func (c *Config) ReadAndWatch() error {
	if len(c.Files) > 0 {
		c.filesChan <- c.Files
		close(c.filesChan)
		return nil
	}

	_, err := ioutil.ReadFile(c.ConfigPath)
	if err != nil {
		return err
	}
	c.stopChan = make(chan struct{})

	go func() {
		//watch the config to update the fswatch if needed
		watcher := fswatcher.StartFileWatcher([]string{c.ConfigPath})
		for {
			select {
			case confChange, ok := <-watcher.GetEventChan():
				if !ok {
					close(c.filesChan)
					fmt.Println("Config watcher stopped")
					return
				}
				// fmt.Printf("ConfEvent %s\n", confChange.Op)
				if len(confChange.Content) > 0 {
					c.Files = strings.Split(string(confChange.Content), "\n")
					c.filesChan <- c.Files
				} else {
					// the config was erased or empty
					c.filesChan <- []string{}
				}
			case <-c.stopChan:
				fmt.Println("Stopping config watcher")
				watcher.Stop()
			}
		}
	}()
	return nil
}
