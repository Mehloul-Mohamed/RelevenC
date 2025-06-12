package scout

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

func Crawl(ctx context.Context, frontier *Frontier) error {
	for {
		url, err := frontier.Pop(ctx)
		if err == redis.Nil {
			time.Sleep(200 * time.Millisecond)
			continue
		} else if err != nil {
			return err
		}

		seen, _ := frontier.TestSeen(ctx, url)
		if seen {
			continue
		}

		out, err := Fetch(ctx, url, frontier)

		var uncrawlable *uncrawlableError
		if errors.As(err, &uncrawlable) {
			frontier.MarkSeen(ctx, url)
			continue
		}

		extracted, err := ExtractUrls(url, out)
		if err != nil {
			continue
		}

		frontier.MarkSeen(ctx, url)
		for _, v := range extracted {
			frontier.Push(ctx, v)
		}
	}
}
