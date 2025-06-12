package scout

import (
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
)

type Frontier struct {
	Client *redis.Client
}

func (f Frontier) Push(ctx context.Context, url string) {
	f.Client.LPush(ctx, "queue", url)
}

func (f Frontier) Pop(ctx context.Context) (string, error) {
	result, err := f.Client.RPop(ctx, "queue").Result()
	if err != nil {
		return "", err
	}
	return result, nil
}

func (f Frontier) MarkSeen(ctx context.Context, url string) {
	f.Client.SAdd(ctx, "seen", url)
}

func (f Frontier) TestSeen(ctx context.Context, url string) (bool, error) {
	result, err := f.Client.SIsMember(ctx, "seen", url).Result()
	if err != nil {
		return false, err
	}
	return result, nil
}
