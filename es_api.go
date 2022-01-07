package elasticsearch

import (
	"context"

	"github.com/tidwall/gjson"
)

// ESClient is an interface which abstracts different
// Elasticsearch versions, providing methods which
// are capable of yielding critical data supporting
// DataSource functionality
type ESClient interface {
	GetShardCount() (int64, error)
	GetDocumentScroller(conf *DataSourceConf, shard int64) ESDocumentScroller
}

// ESDocumentScroller is capable of issuing the configured
// query and scrolling the results, returning strings
// which represent the results of each scroll
type ESDocumentScroller interface {
	IsFinished() bool
	ScrollDocuments(context.Context) (documents []gjson.Result, err error)
}
