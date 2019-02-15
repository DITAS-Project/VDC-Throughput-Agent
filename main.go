package main

import (
	"flag"

	"github.com/DITAS-Project/VDC-Throughput-Agent/throughputagent"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var (
	Build string
)

var logger = logrus.New()
var log *logrus.Entry

func init() {
	if Build == "" {
		Build = "Debug"
	}
	logger.Formatter = new(prefixed.TextFormatter)
	logger.SetLevel(logrus.DebugLevel)
	log = logger.WithFields(logrus.Fields{
		"prefix": "thr-agn",
		"build":  Build,
	})
}

/*
TODO:
	- Logging (like always)
	- push to Elastic
	- failsafe
*/
func main() {

	viper.SetConfigName("traffic")
	viper.AddConfigPath("/etc/ditas/")
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

	log.Infof("config file used @ %s", viper.ConfigFileUsed())

	viper.RegisterAlias("elastic", "ElasticSearchURL")

	flag.String("elastic", "http://localhost:9200", "used to define the elasticURL")
	flag.Int("wt", 1, "wait time for each monitoring window")
	flag.Bool("verbose", false, "enable verbose logging")
	flag.Bool("trace", false, "enable very verbose logging")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	if viper.GetBool("trace") {
		viper.Set("verbose", true)
	}

	if viper.GetBool("verbose") {
		logger.SetLevel(logrus.DebugLevel)
	}

	throughputagent.SetLogger(logger)
	throughputagent.SetLog(log)

	agent, err := throughputagent.NewThroughputAgent()

	if err != nil {
		log.Fatal(err)
	}

	agent.Run()
}
