package util

import (
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
)

// ReadIndexSettingsResponse retrieves the shard count from an index settings response
func ReadIndexSettingsResponse(indexName string, response string) int64 {
	idxName := strings.ReplaceAll(indexName, ".", `\.`)
	idxName = strings.ReplaceAll(idxName, "*", `\*`)
	idxName = strings.ReplaceAll(idxName, "?", `\?`)
	path := fmt.Sprintf("%s.settings.index.number_of_shards", idxName)
	shardCount := gjson.Get(response, path).Int()
	return shardCount
}
