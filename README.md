# Sif ElasticSearch DataSource

An ElasticSearch (6/7) DataSource for Sif.

```bash
$ go get github.com/go-sif/sif-datasource-elasticsearch
```

## Usage

1. Create a `Schema` which represents the fields you intend to extract from each document in the target index:

```go
import (
	"github.com/go-sif/sif"
	"github.com/go-sif/sif/schema"
)

schema := schema.CreateSchema()
schema.CreateColumn("coords.x", &sif.Float64ColumnType{})
schema.CreateColumn("coords.z", &sif.Float64ColumnType{})
schema.CreateColumn("date", &sif.TimeColumnType{Format: "2006-01-02 15:04:05"})
// This datasource will automatically add the following columns to your schema:
//  - es._id (the document id)
//  - es._score (the document score)
```

2. Then, define an ES query to filter data from the target index:

```go
import (
	"github.com/go-sif/sif"
	"github.com/go-sif/sif/schema"
	es7api "github.com/elastic/go-elasticsearch/v7/esapi"
)

// ...
query := "{}" // no need to include index, size or scrolling
			  // params, as they will be overridden by sif
// Full access to the SearchRequest object is provided for further query customization
req := &es7api.SearchRequest{Body: strings.NewReader(query)}
```

3. Finally, define your configuration and create a `DataFrame` which can be manipulated with `sif`:

```go
import (
	"github.com/go-sif/sif"
	"github.com/go-sif/sif/schema"
	esSource "github.com/go-sif/sif-datasource-elasticsearch"
	es7api "github.com/elastic/go-elasticsearch/v7/esapi"
	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
)
// ...
conf := &esSource.DataSourceConf{
	PartitionSize: 128,
	Index:         "my_index_name",
	ScrollTimeout: 10 * time.Minute,
	ES7Query:      req,
	ES7Conf: &elasticsearch7.Config{
		Addresses: []string{"http://1.2.3.4:9200"},
	},
}

dataframe := esSource.CreateDataFrame(conf, schema)
```
