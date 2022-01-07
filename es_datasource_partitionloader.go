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
func (pl *PartitionLoader) Load(parser sif.DataSourceParser) (sif.PartitionIterator, error) {
	return &esPartitionIterator{
		source: pl.source,
		conf:   pl.source.conf,
		shard:  pl.shard,
	}, nil
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
