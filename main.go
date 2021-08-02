package main

import (
	"errors"
	"fmt"
	"github.com/FZambia/sentinel"
	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"log"
	"strings"
	"time"
)

const (
	redisPassword = "master-password"
)

func main() {
	log.SetFlags(log.Flags() | log.Llongfile)
	pool := newSentinelPool()

	conn := pool.Get()

	_ = conn.Send("INCR", "hello")
	_ = conn.Send("INCR", "world")

	res, err := redis.Strings(conn.Do("KEYS", "*"))
	if err != nil {
		log.Println(err)
	}
	log.Println(res)

	// setup jobs

}

func setupWorkers(pool *redis.Pool) {
	var maxConcurrency uint = 10
	var namespace = "my_app"

	workerPool := work.NewWorkerPool(JobContext{}, maxConcurrency, namespace, pool)

	//workerPool.Middleware()

	jobOpts := work.JobOptions{
		Priority:       0,     // default
		MaxFails:       0,     // 0 or 1 => no retry
		SkipDead:       false, // don't store failed jobs
		MaxConcurrency: 0,     // how many workers from the pool will be processing this type of job
	}

	workerPool.JobWithOptions("process_something", jobOpts, (*JobContext).ProcessSomething)

	workerPool.Start()
}

type JobContext struct {
}

func (c *JobContext) ProcessSomething(job *work.Job) {

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

			// auth
			res, err := redis.String(c.Do("AUTH", redisPassword))
			if err != nil {
				log.Println("auth:", err)
				return nil, err
			}
			if res != "OK" {
				log.Println("auth:", res)
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
