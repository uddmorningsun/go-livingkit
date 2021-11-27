package mongodb

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/uddmorningsun/go-livingkit"
	"github.com/uddmorningsun/go-livingkit/config"
	"go.mongodb.org/mongo-driver/bson/mgocompat"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	urllib "net/url"
	"os"
	"time"
)

const (
	mongoAuthenticationDB = "admin"
)

// DefaultMongoClientOptions constructs connection params to the pointer of options.ClientOptions.
// It can also chain call to overwrite default value.
func DefaultMongoClientOptions(mongoCfg config.Connection) *options.ClientOptions {
	var mongoUri = urllib.Values{}
	for key, value := range mongoCfg.Options {
		mongoUri.Set(key, fmt.Sprintf("%s", value))
	}
	authenticationDB, exist := os.LookupEnv(livingkit.MongoAuthenticationDB)
	if !exist {
		authenticationDB = mongoAuthenticationDB
	}
	clientOpts := options.Client().ApplyURI(
		fmt.Sprintf("mongodb://%s/%s?%s", mongoCfg.Address, mongoCfg.Name, mongoUri),
	).SetAuth(options.Credential{
		AuthSource: authenticationDB,
		Username:   mongoCfg.User,
		Password:   mongoCfg.Password,
	})
	return clientOpts
}

// NewMongoConnection initializes connection params with configurations.
func NewMongoConnection(mongoCfg config.Connection, ctx context.Context, opts ...*options.ClientOptions) (*mongo.Client, error) {
	logrus.Infof("initialize db with address: %s, user: %s", mongoCfg.Address, mongoCfg.User)
	if ctx == nil {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		ctx = timeoutCtx
	}
	// Register customized encoder and decoder, details can refer to: bson.Marshal()
	register := mgocompat.NewRegistryBuilder()
	client, err := mongo.Connect(
		ctx, DefaultMongoClientOptions(mongoCfg).SetRegistry(register.Build()), options.MergeClientOptions(opts...),
	)
	if err != nil {
		logrus.Errorf("unable to construct connection for address: %s, error: %s", mongoCfg.Address, err)
		return nil, err
	}
	return client, nil
}
