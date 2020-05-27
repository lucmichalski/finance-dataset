package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
)

const (
	torProxyAddress   = "socks5://51.210.37.251:5566"
	torPrivoxyAddress = "socks5://51.210.37.251:8119"
)

func main() {

	rawUrl := "https://fcpablog.com/2020/05/22/anti-graft-reform-lifts-asia-tempered-by-political-risk/"

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	tbProxyURL, err := url.Parse(torProxyAddress)
	if err != nil {
		log.Fatal(err)
	}
	// pp.Println(tbProxyURL)

	tbDialer, err := proxy.FromURL(tbProxyURL, proxy.Direct)
	if err != nil {
		log.Fatal(err)
	}
	tbTransport := &http.Transport{
		Dial: tbDialer.Dial,
	}
	client.Transport = tbTransport

	// client := new(http.Client)
	request, err := http.NewRequest("GET", rawUrl, nil)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	// log.Println(string(body))
	fmt.Println("Body:", string(body))
	os.Exit(1)

}
