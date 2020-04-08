package elasticsearch

import (
	"bytes"
	ctx "context"
	"fmt"
	"sync"

	elasticsearch6 "github.com/elastic/go-elasticsearch/v6"
	es6api "github.com/elastic/go-elasticsearch/v6/esapi"
	"github.com/go-sif/sif"
	"github.com/go-sif/sif/datasource"
	"github.com/go-sif/sif/datasource/parser/jsonl"
	siferrors "github.com/go-sif/sif/errors"
	"github.com/tidwall/gjson"
)

// with help from https://github.com/elastic/go-elasticsearch/issues/44

type es6PartitionIterator struct {
	source              *DataSource
	shard               int64
	widestInitialSchema sif.Schema
	client              *elasticsearch6.Client
	scrollID            string
	finished            bool
	lock                sync.Mutex
}

func (espi *es6PartitionIterator) OnEnd(onEnd func()) {
	espi.lock.Lock()
	defer espi.lock.Unlock()
}

func (espi *es6PartitionIterator) HasNextPartition() bool {
	espi.lock.Lock()
	defer espi.lock.Unlock()
	return !espi.finished
}

func (espi *es6PartitionIterator) NextPartition() (sif.Partition, error) {
	espi.lock.Lock()
	defer espi.lock.Unlock()
	if !espi.finished && espi.client == nil {
		client, err := elasticsearch6.NewClient(*espi.source.conf.ES6Conf)
		if err != nil {
			return nil, err
		}
		espi.client = client
		espi.source.conf.ES6Query.Index = []string{espi.source.conf.Index}
		espi.source.conf.ES6Query.Preference = fmt.Sprintf("_shards:%d", espi.shard)
		espi.source.conf.ES6Query.Size = &espi.source.conf.PartitionSize
		espi.source.conf.ES6Query.Scroll = espi.source.conf.ScrollTimeout
		res, err := espi.source.conf.ES6Query.Do(ctx.Background(), espi.client)
		if err != nil {
			return nil, err
		}
		if res.IsError() {
			return nil, fmt.Errorf("Unable to scroll documents: %s", res)
		}
		return espi.producePartition(res)
	} else if !espi.finished {
		// otherwise, scroll next document
		res, err := espi.client.Scroll(
			espi.client.Scroll.WithScrollID(espi.scrollID),
			espi.client.Scroll.WithScroll(espi.source.conf.ScrollTimeout),
		)
		if err != nil {
			return nil, fmt.Errorf("Unable to scroll documents: %s", err)
		}
		if res.IsError() {
			return nil, fmt.Errorf("Unable to scroll documents: %s", res)
		}
		return espi.producePartition(res)
	}
	return nil, siferrors.NoMorePartitionsError{}
}

func (espi *es6PartitionIterator) producePartition(res *es6api.Response) (sif.Partition, error) {
	colNames := espi.source.schema.ColumnNames()
	// prefix column names so they search the actual document within the response
	for i, name := range colNames {
		colNames[i] = fmt.Sprintf("_source.%s", name)
	}
	colTypes := espi.source.schema.ColumnTypes()
	defer res.Body.Close()
	var b bytes.Buffer
	b.ReadFrom(res.Body)
	body := b.String()
	espi.scrollID = gjson.Get(body, "_scroll_id").String()
	// check number of results
	hits := gjson.Get(body, "hits.hits").Array()
	// produce partition
	part := datasource.CreateBuildablePartition(espi.source.conf.PartitionSize, espi.widestInitialSchema, espi.source.schema)
	if len(hits) < 1 {
		// close scroll
		espi.client.ClearScroll(espi.client.ClearScroll.WithScrollID(espi.scrollID))
		espi.scrollID = ""
		espi.finished = true
		// return
		return part, nil
	}
	for i := 0; i < len(hits); i++ {
		// create a new row to place values into
		row, err := part.AppendEmptyRowData()
		if err != nil {
			return nil, err
		}
		err = jsonl.ParseJSONRow(colNames, colTypes, hits[i], row)
		if err != nil {
			return nil, err
		}
	}
	return part, nil
}
