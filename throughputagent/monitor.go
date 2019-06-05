//    Copyright 2018 TUB
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package throughputagent

import (
	"context"
	"fmt"
	"math/rand"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	util "github.com/DITAS-Project/TUBUtil"
	"github.com/olivere/elastic"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var logger = logrus.New()
var log = logrus.NewEntry(logger)

func SetLogger(nLogger *logrus.Logger) {
	logger = nLogger
}

func SetLog(entty *logrus.Entry) {
	log = entty
}

type ThroughputAgent struct {
	ElasticSearchURL string
	windowTime       int
	Interface        string
	outChan          chan string
	VDCName          string
	filter           []*regexp.Regexp
	componentNames   []*ComponentMatcher
	elastic          *elastic.Client
	ctx              context.Context
}

type TrafficMessage struct {
	Timestamp time.Time `json:"@timestamp,omitempty"`
	Component string    `json:"traffic.component,omitempty"`
	Send      int       `json:"traffic.send,omitempty"`
	Recived   int       `json:"traffic.recived,omitempty"`
	Total     int       `json:"traffic.total,omitempty"`
}

func NewThroughputAgent() (*ThroughputAgent, error) {

	//compile filter
	ignoreTemplates := viper.GetStringSlice("ignore")
	filter := make([]*regexp.Regexp, 0)
	for _, tpl := range ignoreTemplates {
		log.Debugf("parsing template %s", tpl)
		r, err := regexp.Compile(tpl)
		if err != nil {
			log.Warnf("Ignoring filter template %s because %+v", tpl, err)
		}
		filter = append(filter, r)
	}

	cmp := make([]*ComponentMatcher, 0)
	for k, v := range viper.GetStringMapString("components") {
		m, err := newMatcher(k, v)
		if err != nil {
			log.Warnf("Ignoring filter template %s because %+v", k, err)
		} else {
			cmp = append(cmp, m)
		}

	}

	ta := &ThroughputAgent{
		ElasticSearchURL: viper.GetString("ElasticSearchURL"),
		windowTime:       viper.GetInt("windowTime"),
		VDCName:          viper.GetString("VDCName"),
		componentNames:   cmp,
		filter:           filter,
		outChan:          make(chan string),
		ctx:              context.Background(),
	}

	log.Infof("traffic monitor running with es@%s name:%s waittime:%d", ta.ElasticSearchURL, ta.VDCName, ta.windowTime)

	var client *elastic.Client
	var err error

	if viper.GetBool("ElasticBasicAuth") {
		err = util.WaitForAvailibleWithAuth(ta.ElasticSearchURL, []string{viper.GetString("ElasticUser"), viper.GetString("ElasticPassword")}, nil)

		client, err = elastic.NewClient(
			elastic.SetURL(ta.ElasticSearchURL),
			elastic.SetSniff(false),
			elastic.SetBasicAuth(viper.GetString("ElasticUser"), viper.GetString("ElasticPassword")),
		)
	} else {
		err = util.WaitForAvailible(ta.ElasticSearchURL, nil)

		client, err = elastic.NewClient(
			elastic.SetURL(ta.ElasticSearchURL),
			elastic.SetSniff(false),
		)
	}

	if err != nil {
		log.Error("failed to connect to elastic serach")
		return nil, err
	}

	ta.elastic = client

	return ta, nil
}

func (ta *ThroughputAgent) Run() {

	//util.WaitForGreen(ta.elastic, nil)

	go ta.pktstat()

	for {
		select {
		case out := <-ta.outChan:
			ts := time.Now()

			sample := false
			if viper.GetBool("trace") {
				sample = true
			} else if viper.GetBool("verbose") {
				if rand.Float32() < 0.4 {
					sample = true
				}
			}

			bulkInsert := CreateBulkInsert(ta.elastic, ta.VDCName)

			for ip, data := range ta.readStats(out) {
				if sample {
					log.Debugf("got data from %s with %d", ip, (data[0] + data[1]))
				}

				if ta.filter != nil && len(ta.filter) > 0 {
					skip := false
					for _, m := range ta.filter {
						if m.MatchString(ip) {
							skip = true
							break
						}
					}
					if skip {
						log.Debugf("ignoring %s", ip)
						continue
					}
				}

				comp := ta.getComponentName(ip)

				msg := TrafficMessage{
					Timestamp: ts,
					Component: comp,
					Send:      data[0],
					Recived:   data[1],
					Total:     data[0] + data[1],
				}
				InsertIntoElastic(msg, bulkInsert)
			}

			_, err := bulkInsert.Do(ta.ctx)

			if err != nil {
				log.Error("cound not push to es", err)
			} else {
				log.Debug("pushed new traffic data to es")
			}

			go ta.pktstat()
		}
	}
}

func (ta *ThroughputAgent) readStats(out string) map[string][]int {
	lines := strings.Split(out, "\n")
	stats := make(map[string][]int)
	for _, line := range lines {
		data := strings.Split(line, " ")

		if len(data) == 6 {
			bytes, err := strconv.Atoi(data[0])
			if err != nil {
				bytes = 0
			}
			tx := data[3]
			rx := data[5]

			if _, ok := stats[tx]; !ok {
				stats[tx] = make([]int, 2)
			}
			stats[tx][0] += bytes

			if _, ok := stats[rx]; !ok {
				stats[rx] = make([]int, 2)
			}
			stats[rx][1] += bytes
		}
	}
	return stats
}

func (ta *ThroughputAgent) pktstat() {
	args := make([]string, 0)

	//run only once and terminate
	args = append(args, "-1")

	//output Bytes per second
	args = append(args, "-B")

	//no hostnames
	args = append(args, "-n")

	//waittime
	args = append(args, "-w")
	args = append(args, fmt.Sprintf("%d", ta.windowTime))

	//specific interface
	if ta.Interface != "" {
		args = append(args, "-i")
		args = append(args, ta.Interface)
	}

	//run function
	out, err := exec.Command("pktstat", args...).Output()
	if err != nil {
		log.Errorf("failed to run pktstats, are you root? %+v\n", err)
	}
	ta.outChan <- fmt.Sprintf("%s", out)
}

func (ta *ThroughputAgent) getComponentName(ip string) string {
	for _, m := range ta.componentNames {
		if m.Match(ip) {
			return m.Name
		}
	}
	return ip
}

type ComponentMatcher struct {
	matcher *regexp.Regexp
	Name    string
}

func newMatcher(tpl string, name string) (*ComponentMatcher, error) {

	r, err := regexp.Compile(tpl)
	if err != nil {
		log.Warnf("Ignoring filter template %s because %+v", tpl, err)
		return nil, err
	}

	return &ComponentMatcher{
		matcher: r,
		Name:    name,
	}, nil
}

func (m *ComponentMatcher) Match(ip string) bool {
	return m.matcher.MatchString(ip)
}

//CreateBulkInsert creates a BulkIndexInsertService for trafficData
func CreateBulkInsert(client *elastic.Client, vdcName string) *elastic.BulkService {
	bulkInsert := client.Bulk().Index(util.GetElasticIndex(vdcName)).Type("data")
	return bulkInsert
}

//InsertIntoElastic Adds a TrafficMessage into the bulkInsert
func InsertIntoElastic(msg TrafficMessage, bulkInsert *elastic.BulkService) {
	req := elastic.NewBulkIndexRequest().Doc(msg)
	bulkInsert.Add(req)
}
