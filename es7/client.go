package es7

import (
	"bytes"
	"fmt"
	"log"

	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	es7api "github.com/elastic/go-elasticsearch/v7/esapi"
	sifes "github.com/go-sif/sif-datasource-elasticsearch"
	"github.com/go-sif/sif-datasource-elasticsearch/internal/util"
	"github.com/jinzhu/copier"
	"github.com/tidwall/gjson"
)

type es7client struct {
	api   *elasticsearch7.Client
	query *es7api.SearchRequest
	index string
}

func (c *es7client) GetShardCount() (int64, error) {
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

func (c *es7client) GetDocumentScroller(conf *sifes.DataSourceConf, shard int64) sifes.ESDocumentScroller {
	queryCopy := &es7api.SearchRequest{}
	err := copier.CopyWithOption(queryCopy, c.query, copier.Option{IgnoreEmpty: true, DeepCopy: true})
	if err != nil {
		log.Fatal("unable to deep copy ES query")
	}
	return &es7scroller{
		client:    c,
		conf:      conf,
		shard:     shard,
		queryCopy: queryCopy,
	}
}

// CreateClient returns a new ESClient for Elasticsearch v6
func CreateClient(conf *elasticsearch7.Config, query *es7api.SearchRequest, index string) (sifes.ESClient, error) {
	client, err := elasticsearch7.NewClient(*conf)
	if err != nil {
		return nil, err
	}

	if len(index) == 0 {
		log.Fatal("Must specify an index name")
	}

	// set up query
	query.Index = []string{index}

	return &es7client{client, query, index}, nil
}
