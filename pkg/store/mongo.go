package store

import mongoDriver "go.mongodb.org/mongo-driver/mongo"

var (
	client *mongoDriver.Client
	dbName string
)

func InitMongo(c *mongoDriver.Client, db string) {
	client = c
	dbName = db
}

func Collection(name string) *mongoDriver.Collection {
	if client == nil {
		panic("mongo store not initialized")
	}
	return client.Database(dbName).Collection(name)
}
