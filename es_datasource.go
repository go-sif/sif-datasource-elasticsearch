package elasticsearch

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	elasticsearch6 "github.com/elastic/go-elasticsearch/v6"
	es6api "github.com/elastic/go-elasticsearch/v6/esapi"
	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	es7api "github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/go-sif/sif"
	"github.com/go-sif/sif/datasource"
	"github.com/tidwall/gjson"
)

// DataSource is an ElasticSearch index containing documents which will be manipulating according to a DataFrame
type DataSource struct {
	schema sif.Schema
	conf   *DataSourceConf
}

// DataSourceConf configures an ElasticSearch DataSource
type DataSourceConf struct {
	PartitionSize int
	Index         string
	ScrollTimeout time.Duration
	ES6Query      *es6api.SearchRequest
	ES7Query      *es7api.SearchRequest
	ES6Conf       *elasticsearch6.Config
	ES7Conf       *elasticsearch7.Config
}

// CreateDataFrame is a factory for DataSources
func CreateDataFrame(conf *DataSourceConf, schema sif.Schema) sif.DataFrame {
	var source sif.DataSource
	if conf.PartitionSize == 0 {
		conf.PartitionSize = 128
	}
	if conf.ScrollTimeout == 0 {
		conf.ScrollTimeout = time.Minute * 10
	}
	if conf.ES6Conf != nil && conf.ES7Conf != nil {
		log.Fatal("Cannot specify ES6Conf and ES7Conf simultaneously")
	} else if conf.ES6Conf == nil && conf.ES7Conf == nil {
		log.Fatal("Must specify ES6Conf or ES7Conf")
	} else if conf.ES6Query != nil && conf.ES7Query != nil {
		log.Fatal("Cannot specify ES6Query and ES7Query simultaneously")
	} else if conf.ES6Query == nil && conf.ES7Query == nil {
		log.Fatal("Must specify ES6Query or ES7Query")
	} else if conf.ES6Conf != nil && conf.ES7Query != nil {
		log.Fatal("Must specify an ES6Query with an ES6Conf")
	} else if conf.ES7Conf != nil && conf.ES6Query != nil {
		log.Fatal("Must specify an ES7Query with an ES7Conf")
	} else if len(conf.Index) == 0 {
		log.Fatal("Must specify an Index name")
	} else {
		// add ES-specific fields to schema
		schema.CreateColumn("es._id", &sif.StringColumnType{Length: 512})
		schema.CreateColumn("es._score", &sif.Float32ColumnType{})
		source = &DataSource{schema, conf}
		df := datasource.CreateDataFrame(source, nil, schema)
		return df
	}
	return nil
}

// Analyze returns a PartitionMap, describing how the source file will be divided into Partitions
func (es *DataSource) Analyze() (sif.PartitionMap, error) {
	var shardCount int64
	var body string
	// get the number of shards
	if es.conf.ES6Conf != nil {
		client, err := elasticsearch6.NewClient(*es.conf.ES6Conf)
		if err != nil {
			return nil, err
		}
		res, err := client.Indices.GetSettings(
			client.Indices.GetSettings.WithIndex(es.conf.Index),
			client.Indices.GetSettings.WithIgnoreUnavailable(true),
		)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		var b bytes.Buffer
		b.ReadFrom(res.Body)
		body = b.String()
		if res.IsError() {
			errorType := gjson.Get(body, "error.type")
			errorReason := gjson.Get(body, "error.reason")
			log.Fatalf("[%s] %s: %s",
				res.Status(),
				errorType.String(),
				errorReason.String(),
			)
		}
	} else if es.conf.ES7Conf != nil {
		client, err := elasticsearch7.NewClient(*es.conf.ES7Conf)
		if err != nil {
			return nil, err
		}
		res, err := client.Indices.GetSettings(
			client.Indices.GetSettings.WithIndex(es.conf.Index),
			client.Indices.GetSettings.WithIgnoreUnavailable(true),
		)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		var b bytes.Buffer
		b.ReadFrom(res.Body)
		body = b.String()
		if res.IsError() {
			errorType := gjson.Get(body, "error.type")
			errorReason := gjson.Get(body, "error.reason")
			return nil, fmt.Errorf("[%s] %s: %s",
				res.Status(),
				errorType.String(),
				errorReason.String(),
			)
		}
	}
	idxName := strings.ReplaceAll(es.conf.Index, ".", `\.`)
	idxName = strings.ReplaceAll(idxName, "*", `\*`)
	idxName = strings.ReplaceAll(idxName, "?", `\?`)
	path := fmt.Sprintf("%s.settings.index.number_of_shards", idxName)
	shardCount = gjson.Get(body, path).Int()
	return &PartitionMap{shardCount: shardCount, source: es}, nil
}

// DeserializeLoader creates a PartitionLoader for this DataSource from a serialized representation
func (es *DataSource) DeserializeLoader(bytes []byte) (sif.PartitionLoader, error) {
	pl := PartitionLoader{shard: 0, source: es}
	err := pl.GobDecode(bytes)
	if err != nil {
		return nil, err
	}
	return &pl, nil
}

// IsStreaming returns false for ElasticSearch DataSources
func (es *DataSource) IsStreaming() bool {
	return false
}
