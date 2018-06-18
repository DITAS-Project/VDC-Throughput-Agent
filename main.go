package main

import (
	"flag"

	"github.com/DITAS-Project/VDC-Throughput-Agent/throughputagent"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	log "github.com/sirupsen/logrus"
)

/*
TODO:
	- Logging (like always)
	- push to Elastic
	- failsafe
*/
func main() {

	log.SetLevel(log.DebugLevel)

	viper.SetConfigName("traffic")
	viper.AddConfigPath("/.config/")
	viper.AddConfigPath(".config/")
	viper.AddConfigPath(".")

	viper.SetDefault("ElasticSearchURL", "http://localhost:9200")
	viper.SetDefault("windowTime", 1)
	viper.SetDefault("VDCName", "dummyVDC")
	err := viper.ReadInConfig()
	if err != nil {
		log.Errorf("Fatal error config file: %s \n", err)
	}

	viper.RegisterAlias("elastic", "ElasticSearchURL")

	flag.String("elastic", "http://localhost:9200", "used to define the elasticURL")
	flag.Int("wt", 1, "wait time for each monitoring window")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
	agent, err := throughputagent.NewThroughputAgent()

	if err != nil {
		log.Fatal(err)
	}

	agent.Run()
}
