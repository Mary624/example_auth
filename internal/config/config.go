package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Env             string
	Port            int
	SigningKey      string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	DBConfig
}

type DBConfig struct {
	Collection string
	DBName     string
	Host       string
}

const (
	envName        = "ENV"
	portName       = "PORT"
	accessTTLName  = "ACCESS_TTL"
	refreshTTLName = "REFRESH_TTL"
	signingKeyName = "SIGNING_KEY"

	collectionName = "COLLECTION"
	dbName         = "DB_NAME"
	hostName       = "HOST"
)

func MustLoad() Config {
	emptyName := ""
	defer emptyNameErr(emptyName)
	env := os.Getenv(envName)
	if env == "" {
		emptyName = envName
		return Config{}
	}
	portStr := os.Getenv(portName)
	port, err := strconv.Atoi(portStr)
	if err != nil {
		emptyName = portName
		return Config{}
	}
	signingKey := os.Getenv(signingKeyName)
	if signingKey == "" {
		emptyName = signingKeyName
		return Config{}
	}
	accessTokenTTLStr := os.Getenv(accessTTLName)
	accessTokenTTL, err := time.ParseDuration(accessTokenTTLStr)
	if err != nil {
		emptyName = accessTTLName
		return Config{}
	}
	refreshTokenTTLStr := os.Getenv(refreshTTLName)
	refreshTokenTTL, err := time.ParseDuration(refreshTokenTTLStr)
	if err != nil {
		emptyName = refreshTTLName
		return Config{}
	}

	return Config{
		Env:             env,
		Port:            port,
		AccessTokenTTL:  accessTokenTTL,
		RefreshTokenTTL: refreshTokenTTL,
		SigningKey:      signingKey,
		DBConfig:        MustLoadDB(),
	}
}

func MustLoadDB() DBConfig {
	emptyName := ""
	defer emptyNameErr(emptyName)
	host := os.Getenv(hostName)
	if host == "" {
		emptyName = hostName
		return DBConfig{}
	}
	collection := os.Getenv(collectionName)
	if collection == "" {
		emptyName = collectionName
		return DBConfig{}
	}
	db := os.Getenv(dbName)
	if db == "" {
		emptyName = dbName
		return DBConfig{}
	}

	return DBConfig{
		Host:       host,
		DBName:     db,
		Collection: collection,
	}
}

func emptyNameErr(emptyName string) {
	if emptyName != "" {
		log.Fatal(fmt.Sprintf("%s is not set", emptyName))
	}
}
