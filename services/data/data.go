package data

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"time"

	"github.com/lib/pq"
	"sariego.dev/cotalker-bot/base"
)

// DB - sql connection pool
var DB *sql.DB

// CachedClient - uses an internal cache to persist results
type CachedClient struct {
	base.Client
	cache cache
}

// NewCachedClient - wraps a client with a cache
func NewCachedClient(delegate base.Client) CachedClient {
	return CachedClient{delegate, cache{}}
}

func init() {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	DB = db
}

// GetChannelInfo - returns channel info, caches it for a day
func (cc CachedClient) GetChannelInfo(id string) (base.ChannelInfo, error) {
	cached, err := cc.cache.getChannelInfo(id)
	if err != nil {
		expired := err == errExpired
		fresh, err := cc.Client.GetChannelInfo(id)
		if err != nil && expired {
			// fallback to cached
			return cached, nil
		}
		// save to cache and return fresh
		cc.cache.saveChannelInfo(fresh)
		return fresh, err
	}
	// cache hit!
	return cached, nil
}

type cache struct{}

var errExpired = errors.New("cached item expired")

func (cache) getChannelInfo(id string) (base.ChannelInfo, error) {
	info := base.ChannelInfo{ID: id}
	var expires time.Time

	err := DB.QueryRow(
		"select name, users, expires from channel_info where channel_id = $1",
		id,
	).Scan(&info.Name, pq.Array(&info.Participants), &expires)

	if err == nil && expires.Before(time.Now()) {
		err = errExpired
	}

	return info, err
}

func (cache) saveChannelInfo(info base.ChannelInfo) error {
	_, err := DB.Exec("insert into channel_info(channel_id, name, users) "+
		"values($1,$2,$3) "+
		"on conflict(channel_id) do update "+
		"set name = excluded.name, users = excluded.users, "+
		"updated = default, expired = default",
		info.ID, info.Name, pq.Array(info.Participants),
	)

	return err
}
