package mongodb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type MongoDB struct {
	client *mongo.Client
	database *mongo.Database
	// collection *mongo.Collection
	logger 	*zap.Logger
	mu sync.RWMutex

}

type MongoPool struct {
	clients []*MongoDB
	logger *zap.Logger
	mu sync.Mutex
}

//create a new mongodb instance
func NewMongoDB(uri, dbName/*, collectionName */string, poolSize int, logger *zap.Logger)(*MongoPool, error){
	pool := &MongoPool{
		clients : make([]*MongoDB, 0, poolSize),
		logger: logger,
	}

	for i := 0; i < poolSize; i++ {
		client, err := newMongoClient(uri, dbName/*, collectionName*/,logger)
		if err != nil {
			return nil, fmt.Errorf("Failed to create MongoDB client: %v",err)
		}
		pool.clients = append(pool.clients, client)
	}

	return pool, nil

}

func newMongoClient(uri, dbName/*, collectionName */string , logger *zap.Logger)(*MongoDB, error){
	clientOptions := options.Client().ApplyURI(uri).
	SetMaxPoolSize(100).
	SetMinPoolSize(10).
	SetConnectTimeout(10 * time.Second).
	SetServerSelectionTimeout(10*time.Second).
	SetSocketTimeout(10*time.Second)

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil{
		return nil, fmt.Errorf("Failed to connect to MongoDB: %v",err)
	}

	database := client.Database(dbName)
	// collection := database.Collection(collectionName)

	return &MongoDB{
		client: client,
		database: database,
		// collection: collection,
		logger: logger,
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

func (p *MongoPool) GetClient()(*MongoDB, error){
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.clients) == 0 {
		p.logger.Error("no clients available in pool")
		return nil, fmt.Errorf("no available clients in the pool")
	}

	client := p.clients[0]
	p.clients = p.clients[1:]

	return client, nil
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

func (p *MongoPool) ReleaseClient(client *MongoDB){
	p.mu.Lock()
	defer p.mu.Unlock()

	p.clients = append(p.clients, client)
}


func (m *MongoDB) InsertOne(ctx context.Context,collectionName string, document interface{})(result *mongo.InsertOneResult, err error){
	err = m.withRetry(func() error {
		m.mu.Lock()
		defer m.mu.Unlock()

		collection := m.GetCollection(collectionName)
		result, err = collection.InsertOne(ctx, document)
		if err != nil {
			return fmt.Errorf("failed to insert document: %v",err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	m.logger.Info("document inserted successfully", zap.Any("document", document))
	return result, nil
}

func (m *MongoDB)InsertOneAsync(ctx context.Context,collectionName string, document interface{}m resultChan chan<- *mongo.InsertOneResult, errChan chan<- error ){
	go func() {
		result,err := m.InsertOne(ctx, collectionName, document)
		if err != nil {
			errChan <- err
		}else {
			resultChan <- result
		}
	}
}

// func (p *MongoPool) Find(ctx context.Context, doc interface{}, result interface{})

func (m *MongoDB)Aggregate(ctx context.Context, collectionName string, pipeline interface{}, result interface{}) error {
	return m.withRetry(func() error {	
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
}

func (m *MongoDB) Find(ctx context.Context, collectionName string, filter interface{}, result interface{}, page, limit int) (error){
	return m.withRetry(func() error {
	
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
}

//disconnect mongodb connection
func (p *MongoPool) Close(ctx context.Context) error {
	for _, client := range p.clients{
		if err := client.client.Disconnect(ctx); err != nil {
			p.logger.Error("failed to disconnect MongoDB client", zap.Error(err))
			return fmt.Errorf("failed to disconnect MongoDB client: %v", err)
		}
		return nil
	}
	return nil
}