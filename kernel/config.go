package kernel

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type Config struct {
	Database    *mongo.Database
	Workers     int
	BuildHash   string
	Environment string
}
