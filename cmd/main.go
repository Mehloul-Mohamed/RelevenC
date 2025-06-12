package cmd

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/mehloul-moahmed/relevanc/config"
	"github.com/mehloul-moahmed/relevanc/internal/scout"
	"github.com/redis/go-redis/v9"
)

func Entrypoint() {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	frontier := scout.Frontier{
		Client: redis.NewClient(&redis.Options{
			Addr: "frontier:6379",
			DB:   0,
		}),
	}

	frontier.Push(ctx, config.StartingUrl)

	for i := 0; i < config.NumWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			err := scout.Crawl(ctx, &frontier)
			if err != nil && !errors.Is(err, context.Canceled) {
				log.Printf("[Worker %d] Fatal: %v", id, err)
				cancel()
			}
		}(i)
	}
	wg.Wait()
}
