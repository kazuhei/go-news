package main

import (
	"io/ioutil"
	"log"
	"os"

	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/nlopes/slack"
)

func main() {
	api := slack.New("")
	os.Exit(run(api))
}

func run(api *slack.Client) int {
	currentChannel, err := getChannel(api)
	if err != nil {
		log.Print(err)
		return 1
	}

	rtm := api.NewRTM()
	timerCh := make(chan bool)
	go rtm.ManageConnection()
	go observeTimer(timerCh)

	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch ev := msg.Data.(type) {
			case *slack.HelloEvent:
				// 入室時に書き込みたい場合はここ

			case *slack.MessageEvent:
				log.Printf("Message: %v\n", ev)
				// ユーザーの書き込みに反応したい場合はここ

			case *slack.InvalidAuthEvent:
				log.Print("Invalid credentials")
				return 1

			}
		case <-timerCh:
			title := getNewsSummary()
			readedTitle, err := readLatestTitle()
			if err != nil {
				log.Print(err)
			}
			if title != readedTitle {
				err := writeLatestTitle(title)
				if err != nil {
					log.Print(err)
				} else {
					rtm.SendMessage(rtm.NewOutgoingMessage(title, currentChannel))
				}
			}
		}
	}
}

func getNewsSummary() string {
	doc, err := goquery.NewDocument("http://www.nikkei.com/markets/kigyo/page/?uah=DF_SEC8_C2_070")
	if err != nil {
		log.Print("url scraping fail")
	}

	var newsTitle = ""
	var path = ""
	var exists = false
	doc.Find("#CONTENTS_MARROW .m-block .cmnc-middle").First().Each(func(_ int, s *goquery.Selection) {
		newsTitle = s.Text()
	})

	doc.Find("#CONTENTS_MARROW .m-block .m-articleTitle_text_link").First().Each(func(_ int, s *goquery.Selection) {
		path, exists = s.Attr("href")
	})

	return newsTitle + " http://www.nikkei.com" + path
}

func observeTimer(tch chan bool) {
	for {
		tch <- true
		time.Sleep(time.Duration(10) * time.Minute)
	}
}

func getChannel(api *slack.Client) (string, error) {
	var currentChannel = ""
	channels, err := api.GetChannels(false)
	if err != nil {
		log.Printf("%s\n", err)
		return currentChannel, err
	}
	for _, channel := range channels {
		if channel.Name == "" && channel.IsMember {
			currentChannel = channel.ID
		}
	}
	return currentChannel, nil
}

func readLatestTitle() (string, error) {
	content, err := ioutil.ReadFile("latest")
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func writeLatestTitle(title string) error {
	content := []byte(title)
	err := ioutil.WriteFile("latest", content, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
