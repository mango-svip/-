package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/geziyor/geziyor"
	"github.com/geziyor/geziyor/exporter"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)


func main() {
	url := "https://www.mzitu.com/"
	geziyor.NewGeziyor(geziyor.Options{
		StartURLs: []string{url},
		ParseFunc: quotesParse,
		Exporters: []geziyor.Exporter{exporter.JSONExporter{}},
		CharsetDetectDisabled:true,
	}).Start()

	group.Wait()
}
var dir int = 0
var group = sync.WaitGroup{}

func quotesParse(r *geziyor.Response) {
	children := r.DocHTML.Find("div.postlist").Find("ul").Children()
	children.Each(func(_ int, s *goquery.Selection) {

		log.Println(fmt.Sprintf("爬取专辑第 %d 套", dir) )

		val, _ := s.Find("a").Attr("href")
		spiderImageList(val, r)

	})

}

func spiderImageList(val string, r *geziyor.Response) {
	group.Add(1)

	r.Geziyor.Get(val, func(resp *geziyor.Response) {

		pic, _ := resp.DocHTML.Find("div.main-image").Find("img").Attr("src")

		point := strings.LastIndex(pic, ".net")
		point2 :=strings.LastIndex(pic,".jpg")

		host := pic[0 : point+4]
		req, _ := http.NewRequest("GET", host, nil)
		req.Header.Set("Referer", host)

		picBase := pic[point+4:point2-2]

		dir += 1;


		go func (){
			// 爬取 1- 50 页 每页一张图
			for i := 1; i< 51 ; i++ {
				if i < 10 {
					req.URL.Path  = fmt.Sprintf(picBase + "0" + "%d.jpg", i);
				} else {
					req.URL.Path  = fmt.Sprintf(picBase + "%d.jpg", i);
				}

				r.Geziyor.Do(&geziyor.Request{Request: req}, downLoad)
			}
			defer group.Done()
		}()







	})
	time.Sleep(500*time.Millisecond)

}


func downLoad(r *geziyor.Response) {

	bytes := r.Body
	log.Println(fmt.Sprintf("%v status code %d", r.Request.URL, r.Response.StatusCode))

	if r.Response.StatusCode != 200 {
		return
	}

	dirName := fmt.Sprintf("%d", dir)
	_, err := os.Stat(dirName)
	exist := os.IsNotExist(err)
	if exist {
		os.Mkdir(dirName, os.ModePerm)
	}
	path := r.Request.URL.Path
	index := strings.LastIndex(path, "/")
	file, _ := os.Create(dirName + "/" + path[index+1:])
	defer file.Close()

	file.Write(bytes)

}