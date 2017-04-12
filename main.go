package main

import (
	"flag"
	"fmt"

	"log"
	"os"

	"time"

	"database/sql"

	fworkerprocessor "github.com/ferrariframework/ferrariworker/processor"
	_ "github.com/ferrariframework/ferrariworker/processor/rabbit"
	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
	"github.com/ottogiron/metricsworker/processor"
	"github.com/ottogiron/metricsworker/worker/rabbit"
)

const adapterFactoryName = "rabbit"

//Processor configurations
var concurrencyFlag int
var waitTimeoutFlag int
var redisAddressFlag string
var redisDBFlag int
var mongoHostFlag string
var mongoEventsDBFlag string

var postgresUserFlag string
var postgresPasswordFlag string
var postgresHostFlag string
var postgresDBFlag string

func init() {
	//Processor init
	flag.IntVar(&concurrencyFlag, "concurrency", 1, "Number of concurrent set of workers running")
	flag.IntVar(&waitTimeoutFlag, "wait-timeout", 500, "Time to wait in miliseconds until new jobs are available in rabbit ")
	flag.StringVar(&redisAddressFlag, "redis-address", "localhost:6379", "Redis address example localhost:6779 ")
	flag.IntVar(&redisDBFlag, "redis-db", 0, "Redis DB ")
	flag.StringVar(&mongoHostFlag, "mongo-host", "localhost", "mongo host localhost")
	flag.StringVar(&mongoEventsDBFlag, "mongo-events-db", "events", "mongo events database")

	flag.StringVar(&postgresUserFlag, "postgres-user", "postgres", "postgres user")
	flag.StringVar(&postgresPasswordFlag, "postgres-password", "mysecret", "postgres password")
	flag.StringVar(&postgresHostFlag, "postgres-host", "localhost", "postgres host")
	flag.StringVar(&postgresDBFlag, "postgres-db", "postgres", "postgres database")

	//initialize adapter available properties
	rabbitConfigurationSchema, err := fworkerprocessor.AdapterSchema(adapterFactoryName)
	if err != nil {
		log.Fatalf("Failed to retrieve configuration schema for %s %s", adapterFactoryName, err)
	}
	for _, property := range rabbitConfigurationSchema.Properties {
		name := adapterFactoryName + "-" + property.Name
		switch property.Type {
		case fworkerprocessor.PropertyTypeString:
			defaultValue := property.Default.(string)
			flag.String(name, defaultValue, property.Description)
		case fworkerprocessor.PropertyTypeInt:
			defaultValue := property.Default.(int)
			flag.Int(name, defaultValue, property.Description)
		case fworkerprocessor.PropertyTypeBool:
			defaultValue := property.Default.(bool)
			flag.Bool(name, defaultValue, property.Description)
		}
	}
}

func main() {
	flag.Parse()

	//Get the processor adapter
	factory, err := fworkerprocessor.AdapterFactory(adapterFactoryName)
	if err != nil {
		log.Printf("Failed to load adapter factory for %s %s", adapterFactoryName, err)
		os.Exit(1)
	}
	adapter := factory.New(rabbitAdapterConfig())

	//Configure tasks processor
	proc := processor.New(
		adapter,
		processor.SetConcurrency(concurrencyFlag),
		processor.SetWaitTimeout(time.Duration(waitTimeoutFlag)),
	)

	//Workers initialization
	//distinctName
	redisClient := redisClient()
	distincNameWorker := rabbit.NewDistincNameWorker(redisClient)

	//hourlyLog
	hourlyLogWorker := rabbit.NewHourlyLogWorker(mongoEventsDBFlag, mongoHostFlag)

	//accountName
	accountNameWorker := rabbit.NewAccountNameWorker(postgresDB())

	//Register workers
	proc.Register("distincName", distincNameWorker)
	proc.Register("hourlyLog", hourlyLogWorker)
	proc.Register("accountName", accountNameWorker)

	//Starts new processor
	log.Printf("Waiting for tasks for %dms", waitTimeoutFlag)
	err = proc.Start()
	if err != nil {
		log.Fatal("Failed to start tasks processor ", err)
	}
}

func postgresDB() *sql.DB {
	connectionString := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable	", postgresUserFlag, postgresPasswordFlag, postgresHostFlag, postgresDBFlag)
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("Failed to open postgres connection %s", err)
	}
	return db
}

func redisClient() *redis.Client {

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddressFlag,
		Password: "",          // no password set
		DB:       redisDBFlag, // use default DB
	})

	_, err := client.Ping().Result()
	if err != nil {
		log.Fatalf("Failed to connect to redis %s", err)
	}
	return client
}

func rabbitAdapterConfig() fworkerprocessor.AdapterConfig {

	//Load all the properties values

	//initialize adapter available properties
	rabbitConfigurationSchema, err := fworkerprocessor.AdapterSchema(adapterFactoryName)

	if err != nil {
		log.Fatalf("Failed to retrieve configuration schema for %s %s", adapterFactoryName, err)
	}
	config := fworkerprocessor.NewAdapterConfig()
	for _, property := range rabbitConfigurationSchema.Properties {
		name := adapterFactoryName + "-" + property.Name
		flag := flag.Lookup(name)
		config.Set(property.Name, flag.Value.String())
	}
	return config
}
