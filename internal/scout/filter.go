package scout

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mehloul-moahmed/relevanc/config"
	"github.com/redis/go-redis/v9"
	"go.qingyu31.com/robotstxt"
)

func GetRobotsTxt(link string) (io.Reader, error) {
	parsedUrl, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s://%s/robots.txt", parsedUrl.Scheme, parsedUrl.Host), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36")

	resp, err := Client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return nil, nil
	}

	return resp.Body, nil
}

func crawlable(ctx context.Context, link string, frontier *Frontier) (bool, string) {
	// Scheme check
	parsedUrl, err := url.Parse(link)
	if err != nil {
		return false, "invalid url"
	}

	if parsedUrl.Scheme != "https" && parsedUrl.Scheme != "http" {
		return false, "bad scheme"
	}

	var robots io.Reader

	cachedRobotstxt, err := frontier.Client.HGet(ctx, "robotstxt:"+parsedUrl.Host, "parsed").Result()
	if err == redis.Nil {
		robots, err = GetRobotsTxt(link)
		if err != nil {
			return false, "Robots.txt error"
		}

		if robots == nil {
			return true, ""
		}

		readRobots, _ := io.ReadAll(robots)
		frontier.Client.HSet(ctx, "robotstxt:"+parsedUrl.Host, "parsed", string(readRobots))
		frontier.Client.Expire(ctx, "robotstxt:"+parsedUrl.Host, time.Hour*24)
		robots = strings.NewReader(string(readRobots))
	} else if err != nil {
		return false, "redis err"
	} else {
		robots = strings.NewReader(cachedRobotstxt)
	}

	matcher := robotstxt.Parse(robots)

	if !matcher.AllowedByRobots([]string{config.UserAgent}, parsedUrl.Path) {
		return false, "blocked by robots.txt"
	}

	lastVisited, err := frontier.Client.HGet(ctx, "visit:"+parsedUrl.Host, "last_visited").Result()

	if err == nil {
		lastVisitedTimestamp, _ := strconv.ParseInt(lastVisited, 10, 64)
		if time.Since(time.Unix(lastVisitedTimestamp, 0)) >= 10*time.Second {
			return true, ""
		} else {
			frontier.Push(ctx, link)
			return false, "Delay"
		}
	}

	// optional: google safebrowsing api for malware detection
	// var conf safebrowsing.Config = safebrowsing.Config{
	// 	APIKey: os.Getenv("googleApiKey"),
	// }
	//
	// browser, err := safebrowsing.NewSafeBrowser(conf)
	// defer browser.Close()

	// if err != nil {
	// 	return false
	// }

	// check, err := browser.LookupURLs([]string{link})
	// if err != nil {
	// 	fmt.Println(err)
	// 	return false
	// }

	// return len(check[0]) == 0

	return true, ""
}
