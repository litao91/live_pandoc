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
	"path/filepath"
)

type MDServer struct {
	host      string
	port      int64
	docPath      string
	pandocCmd string
	csspath string
}

func (server *MDServer) handleReq(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	file := strings.Trim(ps.ByName("file"), "/")
	filePath := path.Join(server.docPath, file)
	fmt.Printf("Loading file: %s\n", filePath)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "file %s not found", filePath)
		return
	}
	if (!strings.HasSuffix(file, ".md")) {
		f, err := os.Open(filePath)
		if err != nil {
			fmt.Fprintf(w, "%v", err)
			return
		}
		defer f.Close()
		data := make([]byte, 4096)
		for {
			data = data[:cap(data)]
			n, err := f.Read(data)
			if err != nil {
				break;
			}
			data = data[:n]
			w.Write(data);
		}
		return
	}

	cmdStr := fmt.Sprintf(server.pandocCmd, server.csspath, filePath)
	fmt.Println(cmdStr)
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

func NewServer(filePath string, port int64, csspath string) (server *MDServer) {
	server = &MDServer{
		host:      "127.0.0.1",
		port:      port,
		docPath:      filePath,
		pandocCmd: "pandoc -s --toc --mathjax=http://cdn.mathjax.org/mathjax/latest/MathJax.js?config=TeX-AMS-MML_HTMLorMML  --from=markdown+pipe_tables --to=html5 -H %s %s",
		csspath: csspath,
	}
	return

}

func main() {
	wd, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	var docPath, cssPath string
	if (len(os.Args) >= 2) {
		docPath =  os.Args[1]
	} else {
		docPath = wd
	}

	cssPath = path.Join(docPath, "pandoc.css")
	server := NewServer(docPath, 3333, cssPath)
	fmt.Printf("%v", server.RunHTTPServer())
}
