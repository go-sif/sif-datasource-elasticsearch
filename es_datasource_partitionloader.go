package elasticsearch

import (
	"encoding/binary"
	"fmt"

	"github.com/go-sif/sif"
)

// PartitionLoader is capable of loading partitions of data from a file
type PartitionLoader struct {
	shard  int64
	source *DataSource
}

// ToString returns a string representation of this PartitionLoader
func (pl *PartitionLoader) ToString() string {
	return fmt.Sprintf("ElasticSearch loader shard: %d", pl.shard)
}

// Load is capable of loading partitions of data from a file
func (pl *PartitionLoader) Load(parser sif.DataSourceParser, widestInitialSchema sif.Schema) (sif.PartitionIterator, error) {
	if pl.source.conf.ES6Conf != nil {
		return &es6PartitionIterator{source: pl.source, shard: pl.shard, widestInitialSchema: widestInitialSchema}, nil
	} else if pl.source.conf.ES7Conf != nil {
		return &es7PartitionIterator{source: pl.source, shard: pl.shard, widestInitialSchema: widestInitialSchema}, nil
	}
	return nil, nil
}

// GobEncode serializes a PartitionLoader
func (pl *PartitionLoader) GobEncode() ([]byte, error) {
	buff := make([]byte, 8)
	binary.LittleEndian.PutUint64(buff, uint64(pl.shard))
	return buff, nil
}

// GobDecode deserializes a PartitionLoader
func (pl *PartitionLoader) GobDecode(in []byte) error {
	pl.shard = int64(binary.LittleEndian.Uint64(in))
	return nil
}
