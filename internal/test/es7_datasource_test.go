package test

import (
	"strings"
	"testing"
	"time"

	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	es7api "github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/go-sif/sif"
	esSource "github.com/go-sif/sif-datasource-elasticsearch"
	"github.com/go-sif/sif-datasource-elasticsearch/es7"
	"github.com/go-sif/sif/coltype"
	"github.com/go-sif/sif/schema"
	"github.com/stretchr/testify/require"
)

func TestElasticSearch7Datasource(t *testing.T) {
	// Create a dataframe for the file, load it, and test things
	id := coltype.VarString("_id")
	score := coltype.Float32("_score")
	name := coltype.VarString("name")
	coordsX := coltype.Float64("coords.x")
	coordsY := coltype.Float64("coords.y")
	coordsZ := coltype.Float64("coords.z")
	date := coltype.Time("date", "2006-01-02 15:04:05")
	schema, err := schema.CreateSchema(id, score, name, coordsX, coordsY, coordsZ, date)
	require.Nil(t, err)

	query := ""

	client, err := es7.CreateClient(
		&elasticsearch7.Config{
			Addresses: []string{"http://0.0.0.0:9200"},
		},
		&es7api.SearchRequest{Body: strings.NewReader(query)},
		"edsm",
	)
	require.Nil(t, err)

	conf := &esSource.DataSourceConf{
		PartitionSize: 128,
		ScrollTimeout: 10 * time.Minute,
		Client:        client,
	}

	dataframe := esSource.CreateDataFrame(conf, schema)

	pm, err := dataframe.GetDataSource().Analyze()
	require.Nil(t, err, "Analyze err should be null")
	totalRows := 0
	for pm.HasNext() {
		pl := pm.Next()
		ps, err := pl.Load(nil)
		require.Nil(t, err)
		for ps.HasNextPartition() {
			part, _, err := ps.NextPartition()
			require.Nil(t, err)
			totalRows += part.GetNumRows()
			part.(sif.CollectedPartition).ForEachRow(func(row sif.Row) error {
				idVal, err := id.From(row)
				require.Nil(t, err)
				require.Len(t, idVal, 20)
				scoreVal, err := score.From(row)
				require.Nil(t, err)
				require.EqualValues(t, 1.0, scoreVal)
				coordsXVal, err := coordsX.From(row)
				require.Nil(t, err)
				require.True(t, coordsXVal != 0)
				coordsYVal, err := coordsY.From(row)
				require.Nil(t, err)
				require.True(t, coordsYVal != 0)
				coordsZVal, err := coordsZ.From(row)
				require.Nil(t, err)
				require.True(t, coordsZVal != 0)
				nameVal, err := id.From(row)
				require.Nil(t, err)
				require.Greater(t, len(nameVal), 0)
				return nil
			})
		}
	}
	require.Equal(t, 1000, totalRows)
	require.False(t, pm.HasNext())
}
