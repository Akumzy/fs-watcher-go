package main

import (
	"encoding/json"
	"log"
	"os"
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
	IgnoreFiles       []string     `json:"ignoreFiles"`
	Duration          int64        `json:"duration"`
	Path              string       `json:"path"`
	Recursive         bool         `json:"recursive"`
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
				startWatching(func(watchedFiles []fileInfo) {
					log.Println(watchedFiles)
					ioIPC.Reply(channel, watchedFiles, nil)
				})
			}
		})
	}()
	ioIPC.Start()
}

func startWatching(onReady func(watchedFiles []fileInfo)) {
	w = watcher.New()
	w.IgnoreHiddenFiles(config.IgnoreHiddenFiles)
	w.FilterOps(config.Filters...)
	if len(config.IgnoreFiles) > 0 {
		w.Ignore(config.IgnoreFiles...)
	}
	if config.Path != "" {
		if config.Recursive {
			w.AddRecursive(config.Path)
		} else {
			w.Add(config.Path)
		}
	}
	// Add Recursively
	ioIPC.OnReceiveAndReply("app:addRecursive", func(channel string, data interface{}) {
		path := data.(string)
		if path != "" {
			if err := w.AddRecursive(path); err != nil {
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
				switch event.Op {
				case watcher.Rename:
				case watcher.Remove:
				case watcher.Move:
					newPath, oldPath := getOldAndNewPath(event.Path)
					file.Path = newPath
					file.OldPath = oldPath
				}
				ioIPC.Send("app:change", eventInfo{Event: event.Op, FileInfo: file})
			case err := <-w.Error:
				ioIPC.Send("app:error", err)
				time.Sleep(time.Second)
				os.Exit(1)
			case <-w.Closed:
				return
			}
		}
	}()
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
	onReady(files)
	if err := w.Start(time.Millisecond * time.Duration(config.Duration)); err != nil {
		ioIPC.Send("app:error", err)
	}
}
func getOldAndNewPath(str string) (path, oldPath string) {
	o := strings.Split(str, "->")
	path = strings.TrimSpace(o[1])
	oldPath = strings.TrimSpace(o[0])
	return
}
