package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"path"
)

type errorBody struct {
	Msg string
	ID  uint64
}

type downloadReq struct {
	FileName  string
	FileUrl   string
	GroupName string
	retry     int
}

const downloadPath = "/home/sduby/qqfiles/"
const downloader = "aria2c"
const channelBuf = 5
const retry = 3

var downloadParam = []string{"$URL", "-d", "$DIR", "-o", "$FILENAME"}
var downloadQueue = make(chan downloadReq, channelBuf)

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	req := downloadReq{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println("cant decode response body")
		return
	}

	req.retry = retry
	downloadQueue <- req
}

func runQueue() {
	for req := range downloadQueue {
		if err := doDownload(&req); err != nil && req.retry != 0 {
			req.retry--
			downloadQueue <- req
		}
	}
}

func doDownload(req *downloadReq) (err error) {
	rootPath := path.Join(downloadPath, req.GroupName)
	param := make([]string, len(downloadParam))
	copy(param, downloadParam)
	for i, v := range param {
		switch v {
		case "$URL":
			param[i] = req.FileUrl
		case "$OUTPUT":
			param[i] = path.Join(rootPath, req.FileName)
		case "$DIR":
			param[i] = rootPath
		case "$FILENAME":
			param[i] = req.FileName
		}
	}

	exec.Command("mkdir", "-p", rootPath).Run()
	err = exec.Command(downloader, param...).Run()
	return err
}

func main() {
	for i := 0; i != channelBuf; i++ {
		go runQueue()
	}
	http.HandleFunc("/remoteDownload", downloadHandler)
	http.ListenAndServe("127.0.0.1:3010", nil)
}
