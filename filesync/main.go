package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
)

var groupNameMap sync.Map

const SERVER = "http://localhost:3010/remoteDownload"
const MAXSIZE = 300 * 1024 * 1024

func getGroupName(id int64, ctx *zero.Ctx) string {
	if name, ok := groupNameMap.Load(id); ok {
		return name.(string)
	}

	group := ctx.GetGroupInfo(id, true)
	groupNameMap.Store(id, group.Name)
	return group.Name
}

func callRemoteDownload(filename string, fileurl string, groupname string) {
	data := map[string]string{
		"FileName":  filename,
		"FileUrl":   fileurl,
		"GroupName": groupname,
	}
	jsonvalue, _ := json.Marshal(data)
	log.Println(string(jsonvalue))
	resp, err := http.Post(SERVER, "application/json", bytes.NewBuffer(jsonvalue))
	if err != nil {
		log.Printf("error calling api %v", err)
	}
	defer resp.Body.Close()
}

func main() {
	zero.OnNotice().Handle(func(ctx *zero.Ctx) {
		file := ctx.Event.File
		if ctx.Event.NoticeType != "group_upload" {
			return
		}

		if file.Size > MAXSIZE {
			return
		}

		params := zero.Params{
			"group_id": ctx.Event.GroupID,
			"file_id":  file.ID,
			"busid":    file.BusID,
		}

		resp := ctx.CallAction("get_group_file_url", params)
		url := resp.Data.Get("url").String()
		groupname := getGroupName(ctx.Event.GroupID, ctx)
		filename := file.Name
		log.Println(filename, url, groupname)
		go callRemoteDownload(filename, url, groupname)
	})

	zero.OnCommand("hello", zero.SuperUserPermission).Handle(func(ctx *zero.Ctx) {
		ctx.Send("world")
	})

	zero.RunAndBlock(zero.Config{
		NickName:      []string{"bot"},
		CommandPrefix: "/",
		SuperUsers:    []string{"2081807694"},
		Driver: []zero.Driver{
			driver.NewWebSocketClient("ws://127.0.0.1:6700", "access_token"),
		},
	})
}
