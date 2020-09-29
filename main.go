package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"

	uuid "github.com/satori/go.uuid"

	"github.com/gocolly/colly"
)

func main() {
	biying := "https://cn.bing.com"

	c := colly.NewCollector()

	reg := regexp.MustCompile(`data-ultra-definition-src="([^"]*)"`)

	// 查找背景元素
	c.OnHTML("#bgDiv", func(e *colly.HTMLElement) {

		html, err := e.DOM.Html()
		if err != nil {
			panic(err)
		}

		//正则匹配 src属性
		result := reg.FindAllStringSubmatch(html, -1)

		if len(result[0]) != 2 {
			return
		}

		imageUrl := fmt.Sprintf("%s%s", biying, result[0][1])

		//下载图片
		resp, err := http.Get(imageUrl)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		var out bytes.Buffer
		var stderr bytes.Buffer

		//获取当前用户名
		cmd := exec.Command("whoami")
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			panic(err)
		}

		//保存壁纸文件
		dirPath := fmt.Sprintf(`/Users/%s/dailywallpaper`, strings.Replace(strings.Replace(out.String(), "\r", "", -1), "\n", "", -1))

		os.MkdirAll(dirPath, os.ModeDir|os.ModePerm)

		fileName := fmt.Sprintf("%s/%s.png", dirPath, uuid.NewV4().String())
		file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, os.ModeAppend|os.ModePerm)
		if err != nil {
			panic(err)
		}

		defer file.Close()

		file.Write(body)

		//执行更改桌面壁纸命令
		cmd = exec.Command("/usr/bin/osascript", "-e", fmt.Sprintf(`tell application "System Events" to set picture of every desktop to "%s"`, fileName))
		cmd.Stdout = &out
		cmd.Stderr = &stderr

		err = cmd.Run()
		if err != nil {
			log.Fatal(err.Error(), stderr.String())
		}
	})

	c.OnRequest(func(r *colly.Request) {
		// fmt.Println("Visiting", r.URL)
	})

	//发起请求
	c.Visit(biying)
}
