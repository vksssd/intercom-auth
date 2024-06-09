package mongodb

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	client *mongo.Client
	database *mongo.Database
	collection *mongo.Collection

	mu sync.RWMutex

}

type MongoPool struct {
	clients []*MongoDB
	mu sync.Mutex
}

//create a new mongodb instance
func NewMongoDB(uri, dbName, collectionName string, poolSize int)(*MongoPool, error){
	pool := &MongoPool{
		clients : make([]*MongoDB, 0, poolSize),
	}

	for i := 0; i < poolSize; i++ {
		client, err := newMongoClient(uri, dbName, collectionName)
		if err != nil {
			return nil, fmt.Errorf("Failed to create MongoDB client: %v",err)
		}
		pool.clients = append(pool.clients, client)
	}

	return pool, nil

}

func newMongoClient(uri, dbName, collectionName string )(*MongoDB, error){
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
	collection := database.Collection(collectionName)

	return &MongoDB{
		client: client,
		database: database,
		collection: collection,
	}, nil

}

func (p *MongoPool) GetClient()(*MongoDB, error){
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.clients) == 0 {
		return nil, fmt.Errorf("no available clients in the pool")
	}

	client := p.clients[0]
	p.clients = p.clients[1:]

	return client, nil

}


func (p *MongoPool) ReleaseClient(client *MongoDB){
	p.mu.Lock()
	defer p.mu.Unlock()

	p.clients = append(p.clients, client)
}


func (m *MongoDB) InsertOne(ctx context.Context, document interface{})(*mongo.InsertOneResult, error){
	m.mu.Lock()
	defer m.mu.Unlock()

	result, err := m.collection.InsertOne(ctx, document)
	if err != nil {
		return nil, fmt.Errorf("failed to insert document: %v",err)
	}
	return result, nil
}

func (m *MongoDB)Aggregate(ctx context.Context, pipeline interface{}, result interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cursor, err := m.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return fmt.Errorf("Failed to aggregate documents: %v",err)
	}

	defer cursor.Close(ctx)

	if err = cursor.All(ctx, result); err != nil {
		return fmt.Errorf("Failed to decode documents: %v ", err)
	}

	return nil
}

func (m *MongoDB) Find(ctx context.Context, filter interface{}) ([]bson.M, error){
	cursor, err := m.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find documents: %v",err)
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err = cursor.All(ctx, &result); err!= nil {
		return nil, fmt.Errorf("failed to decode documents: %v",err)
	}
	return result, nil
}

//disconnect mongodb connection
func (p *MongoPool) Close(ctx context.Context) error {
	for _, client := range p.clients{
		if err := client.client.Disconnect(ctx); err != nil {
			return fmt.Errorf("failed to disconnect MongoDB client: %v", err)
		}
		return nil
	}
	return nil
}