package scout

import (
	"errors"
	"io"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func ExtractUrls(baseUrl string, r io.Reader) ([]string, error) {
	var urls []string
	if r == nil {
		return []string{}, errors.New("nil value")
	}
	tokeninzer := html.NewTokenizer(r)

	for {
		tt := tokeninzer.Next()
		t := tokeninzer.Token()

		if tt == html.ErrorToken {
			if tokeninzer.Err() == io.EOF {
				return urls, nil
			}
			return urls, tokeninzer.Err()
		}

		if t.Data == "a" {
			for _, k := range t.Attr {
				if k.Key == "href" {
					base, err := url.Parse(strings.TrimSpace(baseUrl))
					if err != nil {
						continue
					}

					link, err := url.Parse(strings.TrimSpace(k.Val))
					if err != nil {
						continue
					}

					urls = append(urls, base.ResolveReference(link).String())
					break
				}
			}
		}
	}
}
