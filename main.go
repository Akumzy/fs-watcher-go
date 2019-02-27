package main

import (
	"encoding/json"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/Akumzy/ipc"

	"github.com/radovskyb/watcher"
)

type eventInfo struct {
	Event    watcher.Op `json:"event"`
	FileInfo fileInfo   `json:"fileInfo"`
}
type appEvent struct {
	Data  interface{} `json:"data"`
	Error interface{} `json:"error"`
}
type fileInfo struct {
	Size    int64       `json:"size"`
	ModTime time.Time   `json:"modTime"`
	Path    string      `json:"path"`
	Name    string      `json:"name"`
	OldPath string      `json:"oldPath"`
	IsDir   bool        `json:"isDir"`
	Mode    os.FileMode `json:"chmod"`
}

// Config for the app
type Config struct {
	Filters           []watcher.Op `json:"filters"`
	IgnoreHiddenFiles bool         `json:"ignoreHiddenFiles"`
	IgnoreFiles       []string     `json:"ignorePaths"`
	Interval          int64        `json:"interval"`
	Path              string       `json:"path"`
	Recursive         bool         `json:"recursive"`
	FilterHooks       []filterHook `json:"filterHooks"`
}

// filterHook object expecting fron JavaScript
type filterHook struct {
	Reg         string `json:"reg"`
	UseFullPath bool   `json:"useFullPath"`
}

var (
	ioIPC  *ipc.IPC
	config Config
	w      *watcher.Watcher
)

// main
func main() {
	ioIPC = ipc.New()
	go func() {
		ioIPC.Send("app:ready", nil)
		ioIPC.OnReceiveAndReply("app:start", func(channel string, data interface{}) {
			text := data.(string)
			if err := json.Unmarshal([]byte(text), &config); err != nil {
				ioIPC.Reply(channel, nil, err)
				time.Sleep(time.Second)
				os.Exit(1)
			} else {
				go startWatching(func(watchedFiles []fileInfo) {
					ioIPC.Reply(channel, watchedFiles, nil)
				})
			}
		})
		// Add Recursively
		ioIPC.OnReceiveAndReply("app:addRecursive", func(channel string, data interface{}) {
			path := data.(string)

			if path != "" {
				if err := w.AddRecursive(path); err != nil {
					log.Println(err)
					ioIPC.Reply(channel, nil, err)
				} else {
					ioIPC.Reply(channel, true, nil)
				}

			}
		})
		// add
		ioIPC.OnReceiveAndReply("app:add", func(channel string, data interface{}) {
			path := data.(string)
			if path != "" {
				if err := w.Add(path); err != nil {
					ioIPC.Reply(channel, nil, err)
				} else {
					ioIPC.Reply(channel, true, nil)
				}

			}
		})
		// remove
		ioIPC.OnReceiveAndReply("app:remove", func(channel string, data interface{}) {
			path := data.(string)
			if path != "" {
				if err := w.Remove(path); err != nil {
					ioIPC.Reply(channel, nil, err)
				} else {
					ioIPC.Reply(channel, true, nil)
				}

			}
		})

		// removeRecursive
		ioIPC.OnReceiveAndReply("app:removeRecursive", func(channel string, data interface{}) {
			path := data.(string)
			if path != "" {
				if err := w.RemoveRecursive(path); err != nil {
					ioIPC.Reply(channel, nil, err)
				} else {
					ioIPC.Reply(channel, true, nil)
				}

			}
		})
		// ignore
		ioIPC.OnReceiveAndReply("app:ignore", func(channel string, data interface{}) {
			text := data.(string)
			var paths []string
			if err := json.Unmarshal([]byte(text), &paths); err != nil {
				ioIPC.Reply(channel, false, err)

			}
			if len(paths) > 0 {
				if err := w.Ignore(paths...); err != nil {
					ioIPC.Reply(channel, false, err)
				} else {
					ioIPC.Reply(channel, true, nil)
				}

			}
		})
		// app:getWatchedFiles
		ioIPC.OnReceiveAndReply("app:getWatchedFiles", func(channel string, data interface{}) {
			files := getWatchedFiles()
			log.Println(files)
			ioIPC.Reply(channel, files, nil)
		})

	}()
	ioIPC.Start()
}

func startWatching(onReady func(watchedFiles []fileInfo)) {
	w = watcher.New()
	w.IgnoreHiddenFiles(config.IgnoreHiddenFiles)

	if len(config.Filters) > 0 {
		w.FilterOps(config.Filters...)
	}

	if len(config.IgnoreFiles) > 0 {
		w.Ignore(config.IgnoreFiles...)
	}
	// Add Filter Hooks
	go func() {
		if len(config.FilterHooks) > 0 {
			for _, val := range config.FilterHooks {
				r := regexp.MustCompile(val.Reg)
				w.AddFilterHook(watcher.RegexFilterHook(r, val.UseFullPath))
			}
		}
	}()
	if config.Path != "" {
		if config.Recursive {
			if err := w.AddRecursive(config.Path); err != nil {
				reportError(err, true)
			}
		} else {
			if err := w.Add(config.Path); err != nil {
				reportError(err, true)
			}
		}
	}

	go func() {
		for {
			select {
			case event := <-w.Event:
				if event.IsDir() && event.Op == watcher.Write {
					continue
				}
				file := fileInfo{
					Size:    event.Size(),
					Path:    event.Path,
					Name:    event.Name(),
					ModTime: event.ModTime(),
					Mode:    event.Mode(), IsDir: event.IsDir()}
				switch {
				case watcher.Rename == event.Op || watcher.Move == event.Op:
					newPath, oldPath := getOldAndNewPath(event.Path)
					file.Path = newPath
					file.OldPath = oldPath
					file.Name = path.Base(newPath)
				}
				ioIPC.Send("app:change", eventInfo{Event: event.Op, FileInfo: file})
				log.Printf("%+v", event)
			case err := <-w.Error:
				ioIPC.Send("app:error", err)
				log.Println(err)
				time.Sleep(time.Second)
				os.Exit(1)
			case <-w.Closed:
				return
			}
		}
	}()
	files := getWatchedFiles()
	onReady(files)
	if err := w.Start(time.Millisecond * time.Duration(config.Interval)); err != nil {

		reportError(err, true)
	}
}
func getOldAndNewPath(str string) (path, oldPath string) {
	o := strings.Split(str, "->")
	path = strings.TrimSpace(o[1])
	oldPath = strings.TrimSpace(o[0])
	return
}
func reportError(err error, exit bool) {
	log.Println(err)
	ioIPC.Send("app:error", err)
	if exit {
		os.Exit(1)
	}
}

func getWatchedFiles() []fileInfo {
	var files []fileInfo
	for key, file := range w.WatchedFiles() {
		item := fileInfo{
			Size:    file.Size(),
			Path:    key,
			Name:    file.Name(),
			ModTime: file.ModTime(),
			Mode:    file.Mode(), IsDir: file.IsDir()}
		files = append(files, item)
	}
	return files
}
