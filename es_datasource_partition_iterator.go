package elasticsearch

import (
	ctx "context"
	"sync"

	"github.com/go-sif/sif"
	"github.com/go-sif/sif/datasource"
	"github.com/go-sif/sif/datasource/parser/jsonl"
	siferrors "github.com/go-sif/sif/errors"
	"github.com/tidwall/gjson"
)

// with help from https://github.com/elastic/go-elasticsearch/issues/44

type esPartitionIterator struct {
	source   *DataSource
	conf     *DataSourceConf
	shard    int64
	scroller ESDocumentScroller
	lock     sync.Mutex
	colNames []string
}

func (espi *esPartitionIterator) OnEnd(onEnd func()) {
	espi.lock.Lock()
	defer espi.lock.Unlock()
}

func (espi *esPartitionIterator) HasNextPartition() bool {
	espi.lock.Lock()
	defer espi.lock.Unlock()
	return espi.scroller == nil || !espi.scroller.IsFinished()
}

func (espi *esPartitionIterator) NextPartition() (sif.Partition, func(), error) {
	espi.lock.Lock()
	defer espi.lock.Unlock()

	if espi.scroller == nil {
		espi.scroller = espi.source.conf.Client.GetDocumentScroller(espi.conf, espi.shard)
	}

	if !espi.scroller.IsFinished() {
		documents, err := espi.scroller.ScrollDocuments(ctx.Background())
		if err != nil {
			return nil, nil, err
		}
		return espi.producePartition(documents)
	}
	return nil, nil, siferrors.NoMorePartitionsError{}
}

func (espi *esPartitionIterator) producePartition(documents []gjson.Result) (sif.Partition, func(), error) {
	// produce partition
	part := datasource.CreateBuildablePartition(espi.source.conf.PartitionSize, espi.source.schema)
	tempRow := datasource.CreateTempRow()
	for i := 0; i < len(documents); i++ {
		// create a new row to place values into
		row, err := part.AppendEmptyRowData(tempRow)
		if err != nil {
			return nil, nil, err
		}
		err = jsonl.ParseJSONRow(espi.source.schema.ColumnAccessors(), espi.source.jsonPathPrefixes, espi.source.valueHandlers, documents[i], row)
		if err != nil {
			return nil, nil, err
		}
	}
	return part, nil, nil
}
