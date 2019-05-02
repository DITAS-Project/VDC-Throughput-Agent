package throughputagent

import (
	"context"
	util "github.com/DITAS-Project/TUBUtil"
	"github.com/olivere/elastic"
	"testing"
	"time"
)

func TestInsertIntoElastic(t *testing.T) {
	ElasticSearchURL := "http://localhost:9200"
	vdcName := "test-client"

	log.Debug("Waiting for ElasticSearch")
	var timeout time.Duration = 60000
	err := util.WaitForAvailible(ElasticSearchURL, &timeout)
	if err != nil{
		t.Logf("elastic search unavailible: %+v\n", err)
		t.SkipNow()
	}

	log.Infof("using %s for elastic", ElasticSearchURL)

	client, err := elastic.NewSimpleClient(
		elastic.SetURL(ElasticSearchURL),
	)

	ctx := context.Background()

	if err != nil {
		t.Logf("unable to create elastic client tracer: %+v\n", err)
		t.SkipNow()
	}


	insert := CreateBulkInsert(client,vdcName)

	InsertIntoElastic(TrafficMessage{
		time.Now(),
		"test-component",
		5000,
		500,
		5500,
	}, insert)

	res, err := insert.Do(ctx)

	if err != nil {
		log.Errorf("failed to persist traffic data %+v", err)
		t.Fail()
	}

	_, err = client.Get().Index(util.GetElasticIndex(vdcName)).Type("data").Id(res.Succeeded()[0].Id).Do(ctx)

	if err != nil{
		t.Fail()
	}

}