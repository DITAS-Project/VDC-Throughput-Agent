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
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/DITAS-Project/TUBUtil/util"
	"github.com/olivere/elastic"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type ThroughputAgent struct {
	ElasticSearchURL string
	windowTime       int
	Interface        string
	outChan          chan string
	VDCName          string
	elastic          *elastic.Client
	ctx              context.Context
}

type trafficMessage struct {
	Timestamp time.Time `json:"@timestamp,omitempty"`
	Component string    `json:"traffic.component,omitempty"`
	Send      int       `json:"traffic.send,omitempty"`
	Recived   int       `json:"traffic.recived,omitempty"`
	Total     int       `json:"traffic.total,omitempty"`
}

func NewThroughputAgent() (*ThroughputAgent, error) {

	ta := &ThroughputAgent{
		ElasticSearchURL: viper.GetString("ElasticSearchURL"),
		windowTime:       viper.GetInt("windowTime"),
		VDCName:          viper.GetString("VDCName"),
		outChan:          make(chan string),
		ctx:              context.Background(),
	}

	log.Infof("traffic monitor running with es@%s name:%s waittime:%d", ta.ElasticSearchURL, ta.VDCName, ta.windowTime)

	err := util.WaitForAvailible(ta.ElasticSearchURL, nil)

	client, err := elastic.NewClient(
		elastic.SetURL(ta.ElasticSearchURL),
	)

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

			bulkInsert := ta.elastic.Bulk().Index(util.GetElasticIndex(ta.VDCName)).Type("data")

			for ip, data := range ta.readStats(out) {
				msg := trafficMessage{
					Timestamp: ts,
					Component: ip,
					Send:      data[0],
					Recived:   data[1],
					Total:     data[0] + data[1],
				}
				req := elastic.NewBulkIndexRequest().Doc(msg)
				bulkInsert.Add(req)
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
