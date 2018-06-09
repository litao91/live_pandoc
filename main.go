package main

import (
	"fmt"
	"net/http"
	"path"

	"github.com/julienschmidt/httprouter"
)

type MDServer struct {
	host string
	port int64
	path string
}

func (server *MDServer) handleReq(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	file := ps.ByName("file").Trim("/")
	path.Join(server.path, file)
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
		host: "127.0.0.1",
		port: port,
		path: filePath,
	}
	return

}

func main() {
	server := NewServer("", 3333)
	fmt.Printf("%v", server.RunHTTPServer())
}
