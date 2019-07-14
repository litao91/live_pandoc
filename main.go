package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"text/template"

	"path/filepath"

	"io/ioutil"

	"bytes"
	"github.com/julienschmidt/httprouter"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"io"
)

type MDServer struct {
	host         string
	port         int64
	docPath      string
	htmlTemplate *template.Template
}

type MDContent struct {
	Body  string
	Title string
}

func (server *MDServer) handleReq(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	file := strings.Trim(ps.ByName("file"), "/")
	filePath := path.Join(server.docPath, file)
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

	mdContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file! " + filePath)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	var b byte
	var idx int
	for idx, b = range mdContent {
		if b == '\n' {
			break
		}
	}
	line := string(mdContent[:idx])
	// find the first line
	title := strings.TrimLeft(strings.TrimLeft(line, "#"), " ")

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM, extension.Table, extension.DefinitionList, extension.Footnote, extension.Typographer),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	// todo more control on the parsing process
	var html bytes.Buffer
	if err := md.Convert(mdContent, &html); err != nil {
		fmt.Println(err)
	}

	content := MDContent{
		Body:  html.String(),
		Title: title,
	}

	var buf bytes.Buffer

	server.htmlTemplate.Execute(io.Writer(&buf), content)

	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

func (server *MDServer) RunHTTPServer() (err error) {
	router := httprouter.New()
	router.GET("/*file", server.handleReq)
	addr := fmt.Sprintf("%s:%d", server.host, server.port)
	fmt.Println("Listening to " + addr)
	err = http.ListenAndServe(addr, router)
	return
}

func NewServer(filePath string, port int64, htmlTemplate []byte) (server *MDServer) {
	t, err := template.New("html").Parse(string(htmlTemplate))
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	server = &MDServer{
		host:         "0.0.0.0",
		port:         port,
		docPath:      filePath,
		htmlTemplate: t,
	}
	return

}

func main() {
	wd, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	var docPath string
	if len(os.Args) >= 2 {
		docPath = os.Args[1]
	} else {
		docPath = wd
	}

	htmlTemplatePath := path.Join(docPath, "bf_template.html")
	htmlTemplate, err := ioutil.ReadFile(htmlTemplatePath)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	server := NewServer(docPath, 3333, htmlTemplate)
	fmt.Printf("%v\n", server.RunHTTPServer())
}
