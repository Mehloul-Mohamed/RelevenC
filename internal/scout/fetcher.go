package scout

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/mehloul-moahmed/relevanc/config"
)

type uncrawlableError struct {
	reason string
}

var Client = &http.Client{}

func (e *uncrawlableError) Error() string {
	return e.reason
}

func Fetch(ctx context.Context, link string, frontier *Frontier) (io.Reader, error) {
	ok, reason := crawlable(ctx, link, frontier)
	if !ok {
		return nil, &uncrawlableError{reason: reason}
	}

	parsedUrl, _ := url.Parse(link)

	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", config.UserAgent)

	resp, err := Client.Do(req)
	if err != nil {
		return nil, err
	}

	frontier.Client.HSet(ctx, "visit:"+parsedUrl.Host, "last_visited", time.Now().Unix())
	fmt.Println(link)
	return resp.Body, nil
}
