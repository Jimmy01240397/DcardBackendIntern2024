package redis

import (
    "fmt"
    tm "time"
    "encoding/json"
    "context"
    "sync"

    "github.com/redis/go-redis/v9"

    "github.com/Jimmy01240397/DcardBackendIntern2024/utils/time"
    "github.com/Jimmy01240397/DcardBackendIntern2024/utils/config"
)

var lock *sync.RWMutex
var cache *redis.Client
var ctx context.Context
var cursor uint64

type valueTime struct {
    Exptime time.Time `json:"exptime"`
    Value json.RawMessage `json:"value"`
}

func init() {
    lock = new(sync.RWMutex)
    lock.Lock()
    defer lock.Unlock()
    ctx = context.Background()
    cache = redis.NewClient(&redis.Options{
        Addr: config.RedisURL,
        Password: config.RedisPasswd,
        DB: 0,
    })
    _, err := cache.Ping(ctx).Result()
    if err != nil {
        panic(err)
    }
}

func Set(key string, value any, expiration time.Time) (err error) {
    lock.Lock()
    defer lock.Unlock()
    var data []byte
    data, err = json.Marshal(value)
    if err != nil {
        return
    }
    valuetime := valueTime{
        Exptime: expiration,
        Value: json.RawMessage(data),
    }
    data, err = json.Marshal(valuetime)
    if err != nil {
        return
    }
    expdur := expiration.Sub(time.Now())
    if expdur > tm.Millisecond {
        status := cache.Set(ctx, key, string(data), expdur)
        err = status.Err()
    }
    return
}

func Get(key string, value any) (now time.Time, err error) {
    lock.RLock()
    defer lock.RUnlock()
    var data string
    data, err = cache.Get(ctx, key).Result()
    if err != nil {
        return
    }
    var valuetime valueTime
    err = json.Unmarshal([]byte(data), &valuetime)
    if err != nil {
        return
    }
    now = time.Now()
    if now.After(valuetime.Exptime) {
        cache.Del(ctx, key)
        err = fmt.Errorf("expiration")
        return
    }
    err = json.Unmarshal(valuetime.Value, value)
    return
}

func Scan(match string) (keys []string, err error) {
    lock.RLock()
    defer lock.RUnlock()
    keys, cursor, err = cache.Scan(ctx, cursor, match, 0).Result()
    return
}

func Clear() {
    lock.Lock()
    defer lock.Unlock()
    cache.FlushDB(ctx)
}

func Close() {
    lock.Lock()
    defer lock.Unlock()
    cache.Close()
}

