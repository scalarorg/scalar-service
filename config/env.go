package config

import (
	"log"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type ServerEnv struct {
	ENV             string
	CORS_WHITE_LIST []string

	APP_NAME string `validate:"min=1"`

	PORT string `validate:"number"`

	IsProd    bool
	IsStaging bool
	IsDev     bool
	IsTest    bool

	RELAYER_DB_URI string `validate:"uri"`
	INDEXER_DB_URI string `validate:"uri"`

	OPENOBSERVE_ENDPOINT   string `validate:"url"`
	OPENOBSERVE_CREDENTIAL string `validate:"min=1"`

	BITCOIN_CHAIN_ID string `validate:"min=4"`
}

var Env ServerEnv

func LoadEnvWithPath(path string) {
	err := godotenv.Load(os.ExpandEnv(path))
	if err != nil {
		log.Fatalf("Error loading %s file: %s", path, err)
	}

	loadEnv()
}

func LoadEnv() {
	if os.Getenv("ENV") == "" {
		os.Setenv("ENV", "development")
		err := godotenv.Load(os.ExpandEnv(".env"))
		if err != nil {
			log.Fatalln("Error loading .env file: ", err)
		}
	} else if os.Getenv("ENV") == "test" {
		err := godotenv.Load(os.ExpandEnv(".env.test"))
		if err != nil {
			log.Fatalln("Error loading .env.test file: ", err)
		}
	}

	loadEnv()
}

func loadEnv() {
	rawCORSWhiteList := os.Getenv("CORS_WHITE_LIST")
	var corsWhiteList []string
	if rawCORSWhiteList == "" {
		corsWhiteList = []string{
			"http://localhost:3000",
		}
	} else {
		corsWhiteList = strings.Split(rawCORSWhiteList, ",")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "12345"
	}

	Env = ServerEnv{
		ENV:             os.Getenv("ENV"),
		CORS_WHITE_LIST: corsWhiteList,

		APP_NAME: os.Getenv("APP_NAME"),
		PORT:     port,

		OPENOBSERVE_ENDPOINT:   os.Getenv("OPENOBSERVE_ENDPOINT"),
		OPENOBSERVE_CREDENTIAL: os.Getenv("OPENOBSERVE_CREDENTIAL"),

		RELAYER_DB_URI:   os.Getenv("RELAYER_DB_URI"),
		INDEXER_DB_URI:   os.Getenv("INDEXER_DB_URI"),
		BITCOIN_CHAIN_ID: os.Getenv("BITCOIN_CHAIN_ID"),
	}

	validate := validator.New()
	err := validate.Struct(Env)

	if err != nil {
		panic(err)
	}

	Env.IsProd = Env.ENV == "production"
	Env.IsStaging = Env.ENV == "staging"
	Env.IsDev = Env.ENV == "development" || len(Env.ENV) == 0
	Env.IsTest = Env.ENV == "test"
}
