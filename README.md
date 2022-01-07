# Sif ElasticSearch DataSource

[![GoDoc](https://godoc.org/github.com/go-sif/sif?status.svg)](https://pkg.go.dev/github.com/go-sif/sif-datasource-elasticsearch) [![Go Report Card](https://goreportcard.com/badge/github.com/go-sif/sif-datasource-elasticsearch)](https://goreportcard.com/report/github.com/go-sif/sif-datasource-elasticsearch) ![Tests](https://github.com/go-sif/sif-datasource-elasticsearch/workflows/Tests/badge.svg) [![codecov.io](https://codecov.io/github/go-sif/sif-datasource-elasticsearch/coverage.svg?branch=master)](https://codecov.io/gh/go-sif/sif-datasource-elasticsearch?branch=master)

An ElasticSearch (6/7) DataSource for Sif.

```bash
$ go get github.com/go-sif/sif-datasource-elasticsearch@master
$ go get github.com/elastic/go-elasticsearch/v7@7.x
# or
$ go get github.com/elastic/go-elasticsearch/v6@6.x
```

## Usage

1. Create a `Schema` which represents the fields you intend to extract from each document in the target index:

```go
import (
	"github.com/go-sif/sif"
	"github.com/go-sif/sif/schema"
	"github.com/go-sif/sif/coltype"
)

// Any column starting with an _ will be extracted from the
// hits metadata of the query response, such as ID or score
id := coltype.VarString("_id")
score := coltype.Float32("_score")
// Any column which does not start with an _ will be pulled
// automatically from the actual document (_source)
coordsX := coltype.Float64("coords.x")
coordsZ := coltype.Float64("coords.z")
date := coltype.Time("date", "2006-01-02 15:04:05")
schema, err := schema.CreateSchema(id, score, coordsX, coordsZ, date)
```

2. Then, choose an ES version and configure with a target index and query. Elasticsearch 7 is used here as an example:

```go
import (
	"github.com/go-sif/sif"
	"github.com/go-sif/sif/schema"
	"github.com/go-sif/sif/coltype"
	"github.com/go-sif/sif-datasource-elasticsearch/es7"
	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	es7api "github.com/elastic/go-elasticsearch/v7/esapi"
)

// ...
queryJSON := "" // no need to include index, size or scrolling
				// params, as they will be overridden by sif
client, err := es7.CreateClient(
	&elasticsearch7.Config{
		Addresses: []string{"http://0.0.0.0:9200"},
	},
	// Full access to the SearchRequest struct is provided for full query customization
	&es7api.SearchRequest{Body: strings.NewReader(queryJSON)},
	"edsm",
)
```

3. Finally, define your configuration and create a `DataFrame` which can be manipulated with `sif`:

```go
import (
	"github.com/go-sif/sif"
	"github.com/go-sif/sif/schema"
	"github.com/go-sif/sif/coltype"
	"github.com/go-sif/sif-datasource-elasticsearch/es7"
	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	es7api "github.com/elastic/go-elasticsearch/v7/esapi"
	esSource "github.com/go-sif/sif-datasource-elasticsearch"
)
// ...
conf := &esSource.DataSourceConf{
	PartitionSize: 128,
	ScrollTimeout: 10 * time.Minute,
	Client:        client,
}

dataframe := esSource.CreateDataFrame(conf, schema)
```
