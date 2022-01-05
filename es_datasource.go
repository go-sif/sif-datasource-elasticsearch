package elasticsearch

import (
	"log"
	"time"

	"github.com/go-sif/sif"
	"github.com/go-sif/sif/datasource"
	"github.com/go-sif/sif/datasource/parser/jsonl"
)

// DataSource is an ElasticSearch index containing documents which will be manipulating according to a DataFrame
type DataSource struct {
	schema           sif.Schema
	valueHandlers    []jsonl.JSONValueHandler
	jsonPathPrefixes []*string
	conf             *DataSourceConf
}

// DataSourceConf configures an ElasticSearch DataSource
type DataSourceConf struct {
	PartitionSize int
	ScrollTimeout time.Duration
	// ES6Query      *es6api.SearchRequest
	// ES7Query      *es7api.SearchRequest
	Client ESClient
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
	if conf.Client == nil {
		log.Fatal("must specify Client in DataSourceConf")
	} else {
		// construct value handlers for parsing
		valueHandlers, err := jsonl.BuildJSONValueHandlers(schema)
		if err != nil {
			log.Fatal(err)
		}

		// construct column accessor prefixes
		prefixes := make([]*string, schema.NumColumns())
		colNames := schema.ColumnNames()
		// prefix provided column names so they search within the actual document (_source)
		// we only need to prefix columns which DO NOT start with an underscore
		// anything beginning with an underscore (like _score or _id) is an ES field which
		// exists outside of the _source document, but may still be something the client
		// wants to pull into their dataframe
		sourcePrefix := "_source"
		for i, colName := range colNames {
			if len(colName) > 0 && rune(colName[0]) != '_' {
				prefixes[i] = &sourcePrefix
			}
		}

		source = &DataSource{schema: schema, conf: conf, valueHandlers: valueHandlers, jsonPathPrefixes: prefixes}
		df := datasource.CreateDataFrame(source, nil, schema)
		return df
	}
	return nil
}

// Analyze returns a PartitionMap, describing how the source file will be divided into Partitions
func (es *DataSource) Analyze() (sif.PartitionMap, error) {
	// how many shards does the target index have?
	shardCount, err := es.conf.Client.GetShardCount()
	if err != nil {
		return nil, err
	}
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
