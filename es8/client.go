package es8

import (
	"bytes"
	"fmt"
	"log"

	elasticsearch8 "github.com/elastic/go-elasticsearch/v8"
	es8api "github.com/elastic/go-elasticsearch/v8/esapi"
	sifes "github.com/go-sif/sif-datasource-elasticsearch"
	"github.com/go-sif/sif-datasource-elasticsearch/internal/util"
	"github.com/jinzhu/copier"
	"github.com/tidwall/gjson"
)

type es8client struct {
	api   *elasticsearch8.Client
	query *es8api.SearchRequest
	index string
}

func (c *es8client) GetShardCount() (int64, error) {
	res, err := c.api.Indices.GetSettings(
		c.api.Indices.GetSettings.WithIndex(c.index),
		c.api.Indices.GetSettings.WithIgnoreUnavailable(true),
	)
	if err != nil {
		return -1, err
	}
	defer res.Body.Close()
	var b bytes.Buffer
	_, err = b.ReadFrom(res.Body)
	if err != nil {
		return -1, err
	}
	body := b.String()
	if res.IsError() {
		errorType := gjson.Get(body, "error.type")
		errorReason := gjson.Get(body, "error.reason")
		return -1, fmt.Errorf("[%s] %s: %s",
			res.Status(),
			errorType.String(),
			errorReason.String(),
		)
	}
	return util.ReadIndexSettingsResponse(c.index, body), nil
}

func (c *es8client) GetDocumentScroller(conf *sifes.DataSourceConf, shard int64) sifes.ESDocumentScroller {
	queryCopy := &es8api.SearchRequest{}
	err := copier.CopyWithOption(queryCopy, c.query, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	if err != nil {
		log.Fatal("unable to deep copy ES query")
	}
	return &es8scroller{
		client:    c,
		conf:      conf,
		shard:     shard,
		queryCopy: queryCopy,
	}
}

// CreateClient returns a new ESClient for Elasticsearch v6
func CreateClient(conf *elasticsearch8.Config, query *es8api.SearchRequest, index string) (sifes.ESClient, error) {
	client, err := elasticsearch8.NewClient(*conf)
	if err != nil {
		return nil, err
	}

	if len(index) == 0 {
		log.Fatal("Must specify an index name")
	}

	// set up query
	query.Index = []string{index}

	return &es8client{client, query, index}, nil
}
