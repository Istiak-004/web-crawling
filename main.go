package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var userAgent = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3",
	"Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; Googlebot/2.1; +http://www.google.com/bot.html) Chrome/W.X.Y.Z Safari/537.36",
}

func randomUserAgent() string {
	rand.Seed(time.Now().Unix())
	randomNumber := rand.Int() % len(userAgent)
	return userAgent[randomNumber]
}

func perseHTML(resp http.Response) {}

func getRequest(targetUrl string) (*http.Response, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", targetUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", randomUserAgent())
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
func discoverLinks(resp *http.Response, baseUrl string) []string {

	if resp != nil {
		//goquery.NewDocumentFromResponse(resp)
		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			return nil
		}
		defer resp.Body.Close()
		foundLinks := []string{}
		if doc != nil {
			doc.Find("a").Each(func(i int, s *goquery.Selection) {
				res, exist := s.Attr("href")
				if !exist {
					return
				}
				foundLinks = append(foundLinks, res)
			})
		}
		return foundLinks
	}
	return []string{}
}

func checkRelative(link, baseUrl string) string {
	if strings.HasPrefix(link, "/") {
		return fmt.Sprintf("%s%s", baseUrl, link)
	}
	return link
}
func resolveRelativeLinks(link, baseUrl string) (bool, string) {
	resultHref := checkRelative(link, baseUrl)
	baseParse, err := url.Parse(baseUrl)
	if err != nil {
		return false, ""
	}
	resultParse, err := url.Parse(resultHref)
	if err != nil {
		return false, ""
	}

	if baseParse != nil && resultParse != nil {
		if baseParse.Host == resultParse.Host {
			return true, resultHref
		}
		return false, ""
	}
	return false, ""
}

var tokens = make(chan struct{}, 5)

func Cwarl(targetLink, baseUrl string) []string {
	fmt.Printf("here is the target link -- %s", targetLink)
	tokens <- struct{}{} // use semaphore to control the go routine
	resp, _ := getRequest(targetLink)
	<-tokens
	links := discoverLinks(resp, baseUrl)
	foundUrls := []string{}
	for _, link := range links {
		ok, correctLink := resolveRelativeLinks(link, baseUrl)
		if ok {
			if correctLink != "" {
				foundUrls = append(foundUrls, correctLink)
			}
		}
	}

	// need to work on
	//perseHTML(resp)

	return foundUrls
}

func main() {
	fmt.Println("starting cwarling!!")

	var n int
	n++

	baseDomain := "https://www.theguardian.com"
	workList := make(chan []string)
	go func() { workList <- []string{"https://www.theguardian.com"} }()

	seen := make(map[string]bool)
	for ; n > 0; n-- {
		seenLinks := <-workList

		for _, link := range seenLinks {
			if !seen[link] {
				seen[link] = true
				n++
				go func(link string, baseUrl string) {
					foundLinks := Cwarl(link, baseDomain)
					if foundLinks != nil {
						workList <- foundLinks
					}
				}(link, baseDomain)
			}
		}
	}

}
