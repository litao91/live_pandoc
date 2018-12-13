package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"

	"path/filepath"

	"github.com/julienschmidt/httprouter"
)

type MDServer struct {
	host            string
	port            int64
	docPath         string
	pandocCmd       string
	includeHTMLPath string
}

func (server *MDServer) handleReq(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	file := strings.Trim(ps.ByName("file"), "/")
	filePath := path.Join(server.docPath, file)
	fmt.Printf("Loading file: %s\n", filePath)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "file %s not found", filePath)
		fmt.Printf("file %s not found\n", filePath)
		return
	}
	if !strings.HasSuffix(file, ".md") {
		http.ServeFile(w, r, filePath)
		return
	}

	fd, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error reading file! " + filePath)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	scanner := bufio.NewScanner(fd)
	var title string
	if scanner.Scan() {
		title = strings.TrimLeft(strings.TrimLeft(scanner.Text(), "#"), " ")
	}
	fd.Close()

	cmdStr := fmt.Sprintf(server.pandocCmd, server.includeHTMLPath, filePath)
	if title != "" {
		cmdStr = fmt.Sprintf("%s --title \"%s\"", cmdStr, title)
	}

	fmt.Println(cmdStr)
	cmd := exec.Command("bash", "-c", cmdStr)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("%v", err)
		return
	}

	if err = cmd.Start(); err != nil {
		fmt.Printf("%v", err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

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

func NewServer(filePath string, port int64, includeHTMLPath string) (server *MDServer) {
	server = &MDServer{
		host:            "0.0.0.0",
		port:            port,
		docPath:         filePath,
		pandocCmd:       "pandoc --filter graphviz.py -s --columns=200 --toc --mathjax=http://cdn.mathjax.org/mathjax/latest/MathJax.js?config=TeX-AMS-MML_HTMLorMML  --from=markdown+pipe_tables --to=html5 --no-highlight --template %s %s",
		includeHTMLPath: includeHTMLPath,
	}
	return

}

func main() {
	wd, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	var docPath, includeHTMLPath string
	if len(os.Args) >= 2 {
		docPath = os.Args[1]
	} else {
		docPath = wd
	}

	includeHTMLPath = path.Join(docPath, "pandoc_template.html")
	server := NewServer(docPath, 3333, includeHTMLPath)
	fmt.Printf("%v", server.RunHTTPServer())
}
