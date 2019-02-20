// auth: kunlun
// date: 2019-02-20
// description:
package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sys"
	"syscall"
)

var epoller *sys.Epoll

func wsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Connect to server")
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		return true
	}}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	if err = epoller.Add(conn); err != nil {
		log.Printf("Faild to add connection")
		conn.Close()
	}
}

func Start() {
	log.Printf("start success")
	for {
		connections, err := epoller.Wait()
		if err != nil {
			log.Printf("Faild to sys wait %v", err)
			continue
		}
		for _, conn := range connections {
			if conn == nil {
				break
			}
			_, msg, err := conn.ReadMessage()
			if err != nil {
				if err := epoller.Remove(conn); err != nil {
					log.Printf("Faild to remove %v", err)
				}
			} else {
				log.Printf("msg: %s", string(msg))
			}
		}
	}
}

func main() {
	var rLimit syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}
	rLimit.Cur = rLimit.Max
	log.Printf("RLIMIT: %v", rLimit)
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit); err != nil {
		panic(err)
	}

	go func() {
		if err := http.ListenAndServe("0.0.0.0:6060", nil); err != nil {
			log.Fatalf("pprof failed: %v", err)
		}
		log.Printf("pprof port 6060 start success")
	}()
	var err error
	epoller, err = sys.MakeEpoll()
	if err != nil {
		panic(err)
	}

	go Start()

	http.HandleFunc("/", wsHandler)
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatal(err)
	} else {
		log.Printf("ws server port 8000")
	}
}
