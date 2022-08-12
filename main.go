package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"localizer/internal"
)

//go:embed assets/index.html
var html []byte

func main() {
	var dbDir string
	var fileDir string
	var port int
	var disableLaunchBrowser bool
	flag.StringVar(&fileDir, "f", "file", "存放hjson文件的目录。")
	flag.StringVar(&dbDir, "d", "db", "数据库目录。")
	flag.BoolVar(&disableLaunchBrowser, "D", false, "启动后不访问主页。")
	flag.IntVar(&port, "p", 8888, "端口号。")
	flag.Parse()
	log.Println("启动参数:", os.Args)
	h := internal.Handler{
		BaseDir: fileDir,
		DB:      internal.NewDB(dbDir),
	}

	err := os.MkdirAll(dbDir, os.ModePerm)
	if err != nil {
		log.Fatalf("创建数据库目录 %s 失败 %s", dbDir, err)
	}
	err = os.MkdirAll(fileDir, os.ModePerm)
	if err != nil {
		log.Fatalf("创建hjson文件目录 %s 失败 %s", fileDir, err)
	}
	http.HandleFunc("/download", h.DownloadHandler())
	http.HandleFunc("/meta", h.MetaHandler())
	http.HandleFunc("/list", h.ListHandler())
	http.HandleFunc("/upload", h.UploadHandler())
	http.HandleFunc("/", func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = writer.Write(html)
	})
	addr := fmt.Sprintf("0.0.0.0:%d", port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("服务启动失败:%s", err)
	}
	server := &http.Server{}
	server.SetKeepAlivesEnabled(false)
	if !disableLaunchBrowser {
		go open(port)
	}
	log.Fatalln(server.Serve(l))
}
