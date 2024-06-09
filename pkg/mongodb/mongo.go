package mongodb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/sony/gobreaker"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"go.uber.org/zap"
)


type MongoDB struct {
	client *mongo.Client
	database *mongo.Database
	// collection *mongo.Collection
	logger 	*zap.Logger
	cb   *gobreaker.CircuitBreaker
	mu sync.RWMutex

}

type MongoPool struct {
	clients []*MongoDB
	logger *zap.Logger
	mu sync.Mutex
	uri        string
	dbName     string
	poolSize   int
	healthStop chan struct{}
}

type IdempotencyRecord struct {
	IdempotencyKey string      `bson:"idempotency_key"`
	Result         interface{} `bson:"result"`
	Timestamp      time.Time   `bson:"timestamp"`
}


//create a new mongodb instance
func NewMongoDB(uri, dbName/*, collectionName */string, poolSize int, logger *zap.Logger)(*MongoPool, error){
	pool := &MongoPool{
		clients : make([]*MongoDB, 0, poolSize),
		logger: logger,
		uri:uri,
		dbName: dbName,
		poolSize: poolSize,
		healthStop: make(chan struct{}),
	}

	if err := pool.initPool(); err!= nil {
		return nil, err
	}

	go pool.healthCheck()

	return pool, nil

}


func (p *MongoPool)initPool() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// for i := 0; i < poolSize; i++ {
	// 	client, err := newMongoClient(uri, dbName/*, collectionName*/,logger)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("Failed to create MongoDB client: %v",err)
	// 	}
	// 	pool.clients = append(pool.clients, client)
	// }

	for i:= len(p.clients); i < p.poolSize; i++ {
		client, err := newMongoClient(p.uri, p.dbName/*, collectionName*/,p.logger)
			if err != nil {
				return fmt.Errorf("Failed to create MongoDB client: %v",err)
			}
			p.clients = append(p.clients, client)
	}
	return nil
}



func newMongoClient(uri, dbName/*, collectionName */string , logger *zap.Logger)(*MongoDB, error){
	clientOptions := options.Client().ApplyURI(uri).
	SetMaxPoolSize(100).
	SetMinPoolSize(10).
	// SetReplicaSet("rs0").
    // SetReadPreference(readpref.SecondaryPreferred()).
	SetConnectTimeout(10 * time.Second).
	SetServerSelectionTimeout(10*time.Second).
	SetSocketTimeout(10*time.Second)

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil{
		return nil, fmt.Errorf("Failed to connect to MongoDB: %v",err)
	}

	database := client.Database(dbName)
	// collection := database.Collection(collectionName)


	settings := gobreaker.Settings{
		Name: "MongoDB",
		Timeout: 30*time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > 5
		},
	}

	cb := gobreaker.NewCircuitBreaker(settings)

	return &MongoDB{
		client: client,
		database: database,
		// collection: collection,
		logger: logger,
		cb: cb,
	}, nil

}

func (m *MongoDB) withRetry(operation func() error) error{
	operationWithRetry := func () error {
		if err := operation(); err != nil {
			m.logger.Error("operation failed", zap.Error(err))
		}
		return nil
	}
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.MaxElapsedTime = 2*time.Minute
	return backoff.Retry(operationWithRetry, expBackoff)
}

func (m *MongoDB) withCircuitBreaker(operation func () error) error {
	_, err := m.cb.Execute(func () (interface{}, error) {
		return nil, operation()
	})
	if err!=nil {
		m.logger.Error("operation failed due to circuit breaker", zap.Error(err))
	}
	return nil
}

func (p *MongoPool) GetClient(ctx context.Context)(*MongoDB, error){
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if len(p.clients) == 0 {
		p.logger.Warn("No aclient available in pool, atttempting to reinitialize pool")
		if err := p.initPool(); err != nil {
			return nil, fmt.Errorf("failed to reinitalize pool: %v",err)
		}
		if len(p.clients) == 0 {
			return nil, fmt.Errorf("no available clients in the pool after reinitialization")
		}
	}


	client := p.clients[0]
	p.clients = p.clients[1:]

	//ping client to ensure it's alice
	if err := client.client.Ping(ctx, nil); err != nil {
		p.logger.Error("MongoDB Ping failed", zap.Error(err))
		if err := p.Close(ctx); err != nil {
			client.logger.Error("Failed to disconnect from MongoDB", zap.Error(err))
		}
		return nil, fmt.Errorf("MongoDB client is not available")
	}

	return client, nil
}

func (p *MongoPool) ReleaseClient(client *MongoDB) {
	for i, c := range p.clients {
		if c == client {
			p.mu.Lock()
			defer p.mu.Unlock()
		
			// Return the client to the pool
			p.clients = append(p.clients[:i], p.clients[i+1:]...)
		}
	}
}

// func (m *MongoDB) CreateIndex(ctx context.Context)error{
// 	indexModel := mongo.IndexModel {
// 		Keys: bson.M{"name": 1},
// 		Options: options.Index().SetUnique(true),
// 	}

// 	_, err := m.collection.Indexes().CreateOne(ctx, indexModel)
// 	if err != nil {
// 		return fmt.Errorf("Failed to create index:%v", err)
// 	}

// }

func (m *MongoDB) CreateIndexes(ctx context.Context, collectionName string, indexes []mongo.IndexModel) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	collection := m.database.Collection(collectionName)
	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		m.logger.Sugar().Errorf("failed to create indexes for collection: %s: %v", collectionName, err)
		return fmt.Errorf("failed to create indexes for collection: %s: %v", collectionName, err)
	}
	return nil
}

func (m *MongoDB)GetCollection(collectionName string)*mongo.Collection{
	return m.database.Collection(collectionName)
}



func (m *MongoDB) InsertOne(ctx context.Context,collectionName string, document interface{})(result *mongo.InsertOneResult, err error){
	err = m.withRetry(func() error {
		return m.withCircuitBreaker(func() error {
		m.mu.Lock()
		defer m.mu.Unlock()

		collection := m.GetCollection(collectionName)
		result, err = collection.InsertOne(ctx, document)
		if err != nil {
			return fmt.Errorf("failed to insert document: %v",err)
		}
		return nil
	})
})

	if err != nil {
		return nil, err
	}

	m.logger.Info("document inserted successfully", zap.Any("document", document))
	return result, nil
}

func (m *MongoDB)InsertOneAsync(ctx context.Context,collectionName string, document interface{}, resultChan chan<- *mongo.InsertOneResult, errChan chan<- error ){
	go func() {
		result,err := m.InsertOne(ctx, collectionName, document)
		if err != nil {
			errChan <- err
		}else {
			resultChan <- result
		}
	}()
}

// func (p *MongoPool) Find(ctx context.Context, doc interface{}, result interface{})

func (m *MongoDB)Aggregate(ctx context.Context, collectionName string, pipeline interface{}, result interface{}) error {
	return m.withRetry(func() error {	
		return m.withCircuitBreaker(func() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	collection:= m.GetCollection(collectionName)
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		m.logger.Error("failed to aggregate documents", zap.Error(err))
		return fmt.Errorf("Failed to aggregate documents: %v",err)
	}

	defer cursor.Close(ctx)

	if err = cursor.All(ctx, result); err != nil {
		m.logger.Error("failed to decode documents", zap.Error(err))
		return fmt.Errorf("Failed to decode documents: %v ", err)
	}

	m.logger.Info("aggregation successful", zap.Any("result", result))
	return nil
	})

	})

}

func (m *MongoDB) FindPage(ctx context.Context, collectionName string, filter interface{}, result interface{}, page, limit int) (error){
	return m.withRetry(func() error {
		return m.withCircuitBreaker(func() error {
	
	m.mu.Unlock()
	defer m.mu.Unlock()

	findOptions := options.Find()
	if page > 0 && limit > 0 {
		findOptions.SetSkip(int64((page-1)*limit))
		findOptions.SetLimit(int64(limit))
	} 
	
	collection := m.GetCollection(collectionName)
	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		m.logger.Error("failed to find documents", zap.Error(err))
		return fmt.Errorf("failed to find documents: %v",err)
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, result); err!= nil {
		m.logger.Error("failed to decode documents", zap.Error(err))
		return fmt.Errorf("failed to decode documents: %v",err)
	}
	m.logger.Info("documents found successfully", zap.Any("result", result))
	return nil
})
})
}

func (m *MongoDB) PerfomIdempotentTransaction(ctx context.Context, idepotencyKey string, operation func()(interface{}, error))(interface{}, error){
	idempotencyCollection := m.database.Collection("idempotency_key")
	
	sessionOptions := options.Session().
		SetDefaultReadConcern(readconcern.Majority()).
		SetDefaultWriteConcern(writeconcern.New(writeconcern.WMajority()))
	
		session, err := m.client.StartSession(sessionOptions)
	if err !=  nil {
		m.logger.Error("Failed to start session", zap.Error(err))
		return nil, err
	}
	defer session.EndSession(ctx)


	var result interface{}
	err = mongo.WithSession(ctx, session, func(sessCtx mongo.SessionContext)error{
		var record IdempotencyRecord
		err := idempotencyCollection.FindOne(sessCtx, bson.M{"idempotency_key":idepotencyKey}).Decode(&record)
		if err != nil {
			result = record.Result
			return /*record.Result, */nil
		}else if err != mongo.ErrNoDocuments{
			return fmt.Errorf("failed to check idempotency key: %v", err)
		}

		res, err := operation()
		if err != nil {
			return err
		}

		_, err = idempotencyCollection.InsertOne(sessCtx, IdempotencyRecord{
			IdempotencyKey: idepotencyKey,
			Result: res,
			Timestamp: time.Now(),
		})

		if err != nil {
			return fmt.Errorf("failed to store idempotency key: %v", err)
		}

		result = res
		return nil

		// 	}, options.Transaction().SetWriteConcern(mongo.WriteConcernMajority()))
})

	return result, nil
}	



//disconnect mongodb connection
func (p *MongoPool) Close(ctx context.Context) error {
	//stop health check first as it reinitiates the new client
	close(p.healthStop)

	for _, client := range p.clients{
		if err := client.client.Disconnect(ctx); err != nil {
			p.logger.Error("failed to disconnect MongoDB client", zap.Error(err))
			return fmt.Errorf("failed to disconnect MongoDB client: %v", err)
		}
		return nil
	}
	return nil
}


func (p *MongoPool) healthCheck() {
	ticker := time.NewTicker(1*time.Second)
	defer ticker.Stop()

	for {
		select {
		case <- ticker.C:
			p.mu.Lock()
			defer p.mu.Unlock()
			for _, client := range p.clients {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				if err := client.client.Ping(ctx, nil); err!= nil {
					p.logger.Warn("Client failed health check, removing from pool", zap.Error(err))
					if err := client.client.Disconnect(ctx); err != nil {
						client.logger.Error("Failed to disconnect from MongoDB", zap.Error(err))
					}
					// Remove the failed client
					p.ReleaseClient(client)
				}

			}
			p.mu.Unlock()
		case <- p.healthStop:
			return
		}
	}

}