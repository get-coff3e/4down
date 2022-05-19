package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/schollz/progressbar/v3"
)

func download(href string, netpath string) {
	resp, err := http.Get(href)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	URL := strings.Split(href, "/")
	filename := URL[len(URL)-1]
	netpath_href := string(netpath + URL[len(URL)-1])
	// netpath_href would like something like: boards.4channel.org/g/thread/(threadnumber)/(media).jpg

	file, err := os.Create(netpath_href)
	if os.IsExist(err) {
		log.Printf("%v exists, skipping.", filename)
	} else {
		progress := progressbar.DefaultBytes(resp.ContentLength, filename)
		io.Copy(io.MultiWriter(file, progress), resp.Body)
	}
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()
}

func scrape(_URL_ string, param string) {
	// _URL_ is the URL queried to grab data from
	resp, err := http.Get(_URL_)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	log.Printf("'%v' returns %v\n", _URL_, resp.Status)

	MainURL, err := url.Parse(_URL_)
	if err != nil {
		log.Panicln(err)
	}
	netpath := string(MainURL.Hostname() + MainURL.Path + "/")
	//log.Println(netpath)

	if resp.StatusCode == 200 {
		// netpath would be something like "https://boards.4channel.org/g/thread/76759434"
		// href is something like "http://i.4cdn.org/g/1651444537010"
		os.MkdirAll(netpath, 0755)

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Panicln(err)
		}

		switch param {

		case "img":
			log.Println("Grabbing images")
			fileText := []string{}
			doc.Find("div.fileText a").Each(func(i int, s *goquery.Selection) {
				fileText = append(fileText, s.Text())
			})
			doc.Find("a.fileThumb").Each(func(i int, s *goquery.Selection) {
				// For each item found, get the href link (i.4cdn.org)
				fileThumbURL, ok := s.Attr("href")
				if strings.HasPrefix(fileThumbURL, "//") {
					fileThumbURL = strings.Replace(fileThumbURL, "//", "http://", 1)
				}
				if ok {
					download(fileThumbURL, netpath)
				}
			})

		}
	}
}

func main() {

	param := "img"

	cli_URL := []string{}
	man := `4down
Input any number of thread URLs to download them.

Example:	./4down http://boards.4channel.org/g/thread/86763136


Options:
	-h, --help		    Displays this message

	--img               (Default) Download all images off of a thread
	`

	for i := range os.Args {
		arg := os.Args[i]
		switch {
		case strings.HasPrefix(arg, "http"):
			cli_URL = append(cli_URL, arg)
		case strings.Contains(arg, "--help") || strings.Contains(arg, "-h") || len(os.Args) <= 1:
			fmt.Println(man)
		case strings.Contains(arg, "-img"):
			param = "img"
			//case strings.Contains(arg, "--pack="):
			//packtype := strings.Split(arg, "=")[1]
			//continue
			//case strings.Contains(arg, "--path"):
			// can't be bothered to mess with this yet
			//	continue
		}
	}

	for i := range cli_URL {
		scrape(cli_URL[i], param)
		fmt.Printf("Finished downloading %v\n", cli_URL[i])
	}

}
