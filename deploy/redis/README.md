# Redis configuration

## Run redis instances

```yaml
docker run -d --rm \
--name redis-0 \
--net redis \
-p 6379:6379 \
-v ${PWD}/redis-0:/etc/redis/ \
redis redis-server /etc/redis/redis.conf

docker run -d --rm \
--name redis-1 \
--net redis \
-p 6380:6380 \
-v ${PWD}/redis-1:/etc/redis/ \
redis redis-server /etc/redis/redis.conf

docker run -d --rm \
--name redis-2 \
--net redis \
-p 6381:6381 \
-v ${PWD}/redis-2:/etc/redis/ \
redis redis-server /etc/redis/redis.conf
```

## Run sentinels

```yaml
docker run -d --rm \
--name sentinel-0 \
--net redis \
-p 26379:5000 \
-v ${PWD}/sentinel-0:/etc/redis/ \
redis redis-sentinel /etc/redis/sentinel.conf

docker run -d --rm \
--name sentinel-1 \
--net redis \
-p 26380:5000 \
-v ${PWD}/sentinel-1:/etc/redis/ \
redis redis-sentinel /etc/redis/sentinel.conf

docker run -d --rm \
--name sentinel-2 \
--net redis \
-p 26381:5000 \
-v ${PWD}/sentinel-2:/etc/redis/ \
redis redis-sentinel /etc/redis/sentinel.conf
```