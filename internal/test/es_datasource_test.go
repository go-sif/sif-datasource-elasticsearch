package test

import (
	"strings"
	"testing"
	"time"

	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	es7api "github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/go-sif/sif"
	esSource "github.com/go-sif/sif-datasource-elasticsearch"
	"github.com/go-sif/sif/schema"
	"github.com/stretchr/testify/require"
)

func TestElasticSearchDatasource(t *testing.T) {
	// Create a dataframe for the file, load it, and test things
	schema := schema.CreateSchema()
	schema.CreateColumn("name", &sif.VarStringColumnType{})
	schema.CreateColumn("coords.x", &sif.Float64ColumnType{})
	schema.CreateColumn("coords.y", &sif.Float64ColumnType{})
	schema.CreateColumn("coords.z", &sif.Float64ColumnType{})
	schema.CreateColumn("date", &sif.TimeColumnType{Format: "2006-01-02 15:04:05"})

	query := ""

	req := &es7api.SearchRequest{Body: strings.NewReader(query)}

	conf := &esSource.DataSourceConf{
		PartitionSize: 128,
		Index:         "edsm",
		ScrollTimeout: 10 * time.Minute,
		ES7Query:      req,
		ES7Conf: &elasticsearch7.Config{
			Addresses: []string{"http://0.0.0.0:9200"},
		},
	}

	dataframe := esSource.CreateDataFrame(conf, schema)

	pm, err := dataframe.GetDataSource().Analyze()
	require.Nil(t, err, "Analyze err should be null")
	totalRows := 0
	for pm.HasNext() {
		pl := pm.Next()
		ps, err := pl.Load(nil, schema)
		require.Nil(t, err)
		for ps.HasNextPartition() {
			part, err := ps.NextPartition()
			require.Nil(t, err)
			totalRows += part.GetNumRows()
		}
	}
	require.Equal(t, 1000, totalRows)
	require.False(t, pm.HasNext())
}
