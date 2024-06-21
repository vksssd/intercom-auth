package redis

import (
	// "os"

	"context"
	"encoding/json"
	"errors"
	"time"

	// "github.com/cenkalti/backoff/v4"
	"github.com/go-redis/redis/v8"
	"github.com/sony/gobreaker"
	"github.com/vksssd/intercom-auth/config"
	"github.com/RediSearch/redisearch-go/redisearch"

)

var (
	ErrNotFound = errors.New("key not found in Redis")
	)
	
type ConnectionPool struct {
		pool chan *redis.Client
		checkHealth bool
}

type Client struct {
	Config *config.RedisConfig
	// Client *redis.Client
	Ctx context.Context
	circuit *gobreaker.CircuitBreaker
	search *redisearch.Client
	Pool *ConnectionPool
}


// need to optamisse to use newclient that is implement below instead of redis client again here 
func NewConnectionPool(add string, poolSize int, checkHealth bool)*ConnectionPool {
	p := &ConnectionPool{
		pool: make(chan *redis.Client, poolSize),
		checkHealth: checkHealth,
	}
	for i :=0; i <poolSize; i ++ {
		client := redis.NewClient(&redis.Options{
			Addr: add,
			PoolSize: poolSize,
		})
		p.pool<- client
	}

	if checkHealth {
		go p.healthCheck()
	}

	return p
}


func(p *ConnectionPool) GetConnection() *redis.Client {
	return <- p.pool
}

func (p *ConnectionPool) ReleaseConnection( client *redis.Client) {
	p.pool<-client
}

func NewClient(cfg *config.RedisConfig) (*Client, error) {
	
	opt := &redis.Options{
		Addr: cfg.URL,
		PoolSize: cfg.PoolSize,
		MinIdleConns: 5, // update the cofig
	}
	rdb:= redis.NewClient(opt)

	//optamising searching for full text search
	search := redisearch.NewClient(cfg.URL, "myIndex")

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name: "RedisCircuitBreaker",
		MaxRequests: 5,
		Interval: 30*time.Second,
		Timeout: 5*time.Second,
	})

	// rateLimiter := rate

	return &Client{
		// Client: rdb, // this client is now in pool
		Config: cfg,
		Ctx: context.Background(),
		circuit: cb,
		search: search,
	},nil
}

func (c *Client) Set(key string, value interface{}, expiration time.Duration) error {
	client := c.Pool.GetConnection()
	defer c.Pool.ReleaseConnection(client)
	
	_,err := c.circuit.Execute(func() (interface{}, error) {

		data, err := json.Marshal(value)
		if err != nil {
			return nil,err
		}
		return nil,	client.Set(c.Ctx, key, data, expiration).Err()
	})

	return err
}

func (c *Client) Get(key string, dest interface{}) error {
	client := c.Pool.GetConnection()
	defer c.Pool.ReleaseConnection(client)

	var data []byte
	_,err := c.circuit.Execute(func() (interface{}, error) {
		
		var err error
		data, err = client.Get(c.Ctx,key).Bytes()
		
		if err != nil {
			if errors.Is(err, redis.Nil){
				return nil, ErrNotFound
			}
			}
		return nil, err
	})
		
	if err != nil {
			return err
	}
	
	return json.Unmarshal([]byte(data), dest)
}

func (c *Client)Delete(key string) error {
	client := c.Pool.GetConnection()
	defer c.Pool.ReleaseConnection(client)

	_,err := c.circuit.Execute(func () (interface{}, error){
		return nil,  client.Del(c.Ctx, key).Err()
	})
	return err
}

func(c *Client) Close() error {
	close(c.Pool.pool)
	return nil
}

func (c *Client) Retry(ctx context.Context, operation func(context.Context) error ) error {
	var maxRetries = c.Config.RetryAttempts
	var retries int

	for {
		select {
		case<- ctx.Done():
			return ctx.Err()
		default:
			err := operation(ctx)
			if err == nil {
				return nil
			}

			retries++
			if retries >= maxRetries{
				return errors.New("maximum retry attempts reached")
			}

			backoff := time.Duration(retries*retries*100)*time.Microsecond
			select {
			case <- ctx.Done():
				return ctx.Err()
			case <- time.After(backoff):
			}
		}
	}
}

func (c *Client) Pipeline(ctx context.Context, operations ...func(redis.Pipeliner) error )error{
	client := c.Pool.GetConnection()
	defer c.Pool.ReleaseConnection(client)

	pipe := client.TxPipeline()
	for _, op := range operations {
		if err := op(pipe); err !=nil{
			_ = pipe.Close()
			return err
		}
	}

	_, err := pipe.Exec(ctx)
	return err

}

func ( c *Client) AddToSet(ctx context.Context, key string, memeber interface{}) error {
	client := c.Pool.GetConnection()
	defer c.Pool.ReleaseConnection(client)
	return client.SAdd(ctx, key, memeber).Err()
}

func (c *Client) IsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	client := c.Pool.GetConnection()
	defer c.Pool.ReleaseConnection(client)
	return client.SIsMember(ctx, key, member).Result()
}

func (c *Client) CreateIdempotency(ctx context.Context, key string, expiration time.Duration) error {
	client := c.Pool.GetConnection()
	defer c.Pool.ReleaseConnection(client)
	_, err := client.SetNX(ctx, key, "", expiration).Result()
	return err
}


func (c *Client) DeleteIdempotencyKey(ctx context.Context, key string) error {
	client := c.Pool.GetConnection()
	defer c.Pool.ReleaseConnection(client)
	return client.Del(ctx, key).Err()
}


func (c *Client) ClusterInfo(ctx context.Context)(string, error){
	client := c.Pool.GetConnection()
	defer c.Pool.ReleaseConnection(client)

	return client.ClusterInfo(ctx).Result()
}

func (c *Client) Search(ctx context.Context, q string)([]string, error){

	docs,_, err := c.search.Search(redisearch.NewQuery(q))
	if err!= nil {
		return nil, err
	}

	var result []string
	for _, doc := range docs {
		result = append(result, doc.Id)
	}
	return result, nil
}


func (p *ConnectionPool) healthCheck() {
	for {
		select {
		case <- time.After(time.Minute):
			for client := range p.pool {
				if _, err := client.Ping(context.Background()).Result(); err != nil {
					_ = client.Close()
					newClient := redis.NewClient(client.Options()) // optamise to use own new client method
					p.pool<- newClient
				}else{
					p.pool<-client
				}
			}
		}
	}
}