package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type MDServer struct {
	host      string
	port      int64
	path      string
	pandocCmd string
}

func (server *MDServer) handleReq(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	file := strings.Trim(ps.ByName("file"), "/")
	filePath := path.Join(server.path, file)
	cmdStr := fmt.Sprintf(server.pandocCmd, filePath)
	fmt.Println("Command: " + cmdStr)
	cmd := exec.Command("bash", "-c", cmdStr)
	var stdout, err = cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	if err = cmd.Start(); err != nil {
		fmt.Printf("%v", err)
		return
	}

	reader := bufio.NewReader(stdout)
	for {
		data, err := reader.ReadBytes('\n')
		if err != nil {
			fmt.Printf("%s", err)
			return
		}
		w.Write(data)
	}
}

func (server *MDServer) RunHTTPServer() (err error) {
	router := httprouter.New()
	router.GET("/*file", server.handleReq)
	addr := fmt.Sprintf("%s:%d", server.host, server.port)
	fmt.Println("Listening to " + addr)
	err = http.ListenAndServe(addr, router)
	return
}

func NewServer(filePath string, port int64) (server *MDServer) {
	server = &MDServer{
		host:      "127.0.0.1",
		port:      port,
		path:      filePath,
		pandocCmd: "pandoc -s --mathjax=http://cdn.mathjax.org/mathjax/latest/MathJax.js?config=TeX-AMS-MML_HTMLorMML  --from=markdown+pipe_tables --to=html %s",
	}
	return

}

func main() {
	path := os.Args[1]
	server := NewServer(path, 3333)
	fmt.Printf("%v", server.RunHTTPServer())
}
