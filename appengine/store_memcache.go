package appengine

import (
	"time"

	"bitbucket.org/johnsto/go-passwordless"

	"github.com/gyepisam/mcf"
	_ "github.com/gyepisam/mcf/scrypt"
	"golang.org/x/net/context"
	"google.golang.org/appengine/memcache"
)

type MemcacheStore struct {
	KeyPrefix string
}

func (s MemcacheStore) Store(ctx context.Context, token, uid string, ttl time.Duration) error {
	hashToken, err := mcf.Create(token)
	if err != nil {
		return err
	}

	return memcache.Set(ctx, &memcache.Item{
		Key:        s.KeyPrefix + uid,
		Value:      []byte(hashToken),
		Expiration: ttl,
	})
}

// Exists returns true if a token for the specified user exists. The returned
// `time.Time` will always be set to zero as memcache doesn't provide this
// property for retrieved items.
func (s MemcacheStore) Exists(ctx context.Context, uid string) (bool, time.Time, error) {
	_, err := memcache.Get(ctx, s.KeyPrefix+uid)
	if err == memcache.ErrCacheMiss {
		// No known token for this user
		return false, time.Time{}, nil
	} else {
		// Token exists and is still valid
		return true, time.Time{}, nil
	}
}

func (s MemcacheStore) Verify(ctx context.Context, token, uid string) (bool, error) {
	item, err := memcache.Get(ctx, s.KeyPrefix+uid)
	if err == memcache.ErrCacheMiss {
		// No token in database
		return false, passwordless.ErrTokenNotFound
	} else if err != nil {
		return false, err
	}

	hashedToken := string(item.Value)
	if valid, err := mcf.Verify(token, hashedToken); err != nil {
		// Couldn't validate token
		return false, err
	} else if !valid {
		// Token does not validate against hashed token
		return false, nil
	} else {
		// Token is valid!
		return true, nil
	}
}

func (s MemcacheStore) Delete(ctx context.Context, uid string) error {
	return memcache.Delete(ctx, s.KeyPrefix+uid)
}
