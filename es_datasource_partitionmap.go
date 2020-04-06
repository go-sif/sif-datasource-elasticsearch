package elasticsearch

import "github.com/go-sif/sif"

// PartitionMap is an iterator producing a sequence of PartitionLoaders
type PartitionMap struct {
	shardCount   int64
	currentShard int64
	source       *DataSource
}

// HasNext returns true iff there is another PartitionLoader remaining
func (pm *PartitionMap) HasNext() bool {
	return pm.currentShard < pm.shardCount
}

// Next returns the next PartitionLoader for a file
func (pm *PartitionMap) Next() sif.PartitionLoader {
	result := &PartitionLoader{shard: pm.currentShard, source: pm.source}
	pm.currentShard++
	return result
}
