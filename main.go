package main

import (
	"flag"

	"log"
	"os"

	"time"

	fworkerprocessor "github.com/ferrariframework/ferrariworker/processor"
	_ "github.com/ferrariframework/ferrariworker/processor/rabbit"
	"github.com/ottogiron/metricsworker/processor"
	"github.com/ottogiron/metricsworker/worker/rabbit"
)

const adapterFactoryName = "rabbit"

//Processor configurations
var concurrencyFlag int
var waitTimeoutFlag int

//Rabbit configurations
var uriKeyFlag string
var routingKeyFlag string
var bindingKeyFlag string

// queue flags
var queueNameFlag string
var queueDurableFlag bool
var queueDeleteWhenUsedFlag bool
var queueExclusiveFlag bool
var queueNowaitFlag bool

func init() {
	//Processor init
	flag.IntVar(&concurrencyFlag, "concurrency", 1, "Number of concurrent set of workers running")
	flag.IntVar(&waitTimeoutFlag, "wait-timeout", 500, "Time to wait in miliseconds until new jobs are available in rabbit ")

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

	//Register workers
	proc.Register("distincName", &rabbit.DistinctNameWorker{})

	//Starts new processor
	log.Printf("Waiting for tasks for %dms", waitTimeoutFlag)
	err = proc.Start()
	if err != nil {
		log.Fatal("Failed to start tasks processor ", err)
	}
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
