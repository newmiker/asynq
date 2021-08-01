package main

import (
	"errors"
	"fmt"
	"github.com/FZambia/sentinel"
	"github.com/gomodule/redigo/redis"
	"log"
	"strings"
	"time"
)

func main() {
	pool := newSentinelPool()

	conn := pool.Get()
	res, err := redis.String(conn.Do("keys", "*"))
	if err != nil {
		log.Println(err)
	}
	log.Println(res)
}

func newSentinelPool() *redis.Pool {
	sntnl := &sentinel.Sentinel{
		Addrs:      []string{":26379", ":26380", ":26381"},
		MasterName: "mymaster",
		Dial: func(addr string) (redis.Conn, error) {
			timeout := 500 * time.Millisecond
			c, err := redis.Dial("tcp", addr,
				redis.DialConnectTimeout(timeout),
				redis.DialReadTimeout(timeout),
				redis.DialWriteTimeout(timeout))
			if err != nil {
				return nil, err
			}
			return c, nil
		},
	}
	return &redis.Pool{
		MaxIdle:     3,
		MaxActive:   64,
		Wait:        true,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			masterAddr, err := sntnl.MasterAddr()
			if err != nil {
				return nil, err
			}

			// trick against NATs
			parts := strings.Split(masterAddr, ":")
			if len(parts) != 2 {
				err = fmt.Errorf("can not extract port from %s", masterAddr)
				return nil, err
			}
			masterAddr = fmt.Sprintf(":%s", parts[1])

			c, err := redis.Dial("tcp", masterAddr)
			if err != nil {
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if !sentinel.TestRole(c, "master") {
				return errors.New("Role check failed")
			} else {
				return nil
			}
		},
	}
}
