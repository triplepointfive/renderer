package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/websocket"

	"github.com/radovskyb/watcher"
)

var client *websocket.Conn

func echoHandler(ws *websocket.Conn) {
	client = ws
	io.Copy(ws, ws)
}

func main() {
	w := watcher.New()

	http.Handle("/echo", websocket.Handler(echoHandler))
	http.Handle("/", http.FileServer(http.Dir(".")))

	// w.FilterOps(watcher.Write)

	go func() {
		for {
			select {
			case <-w.Event:
				log.Println("Changed")
				if client != nil {
					client.Write([]byte("Reload"))
					client.Close()
					client = nil
				}
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	if err := w.Add("hello.png"); err != nil {
		log.Fatalln(err)
	}

	for path, f := range w.WatchedFiles() {
		fmt.Printf("%s: %s\n", path, f.Name())
	}

	go func() {
		// Start the watching process - it'll check for changes every 100ms.
		if err := w.Start(time.Millisecond * 100); err != nil {
			log.Fatalln(err)
		}
	}()

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
