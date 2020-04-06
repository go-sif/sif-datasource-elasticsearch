package elasticsearch

import (
	"fmt"
	"time"

	"github.com/go-sif/sif"
	"github.com/tidwall/gjson"
)

func parseValue(val gjson.Result, colName string, colType sif.ColumnType, row sif.Row) error {
	// parse type
	switch colType.(type) {
	// TODO array/slice type
	case *sif.BoolColumnType:
		row.SetBool(colName, val.Bool())
	case *sif.Int8ColumnType:
		row.SetInt8(colName, int8(val.Int()))
	case *sif.Int16ColumnType:
		row.SetInt16(colName, int16(val.Int()))
	case *sif.Int32ColumnType:
		row.SetInt32(colName, int32(val.Int()))
	case *sif.Int64ColumnType:
		row.SetInt64(colName, int64(val.Int()))
	case *sif.Float32ColumnType:
		row.SetFloat32(colName, float32(val.Float()))
	case *sif.Float64ColumnType:
		row.SetFloat64(colName, val.Float())
	case *sif.StringColumnType:
		row.SetString(colName, val.String())
	case *sif.TimeColumnType:
		format := colType.(*sif.TimeColumnType).Format
		tval, err := time.Parse(format, val.String())
		if err != nil {
			return fmt.Errorf("Column %s could not be parsed as datetime with format %s. Was: %#v", colName, format, val)
		}
		row.SetTime(colName, tval)
	case *sif.VarStringColumnType:
		row.SetVarString(colName, val.String())
	default:
		return fmt.Errorf("ElasticSearch document parsing does not support column type %T", colType)
	}
	return nil
}

// Parses a slice of strings into a Row, according to a schema
func scanRow(names []string, types []sif.ColumnType, rowJSON gjson.Result, row sif.Row) error {
	for idx, colName := range row.Schema().ColumnNames() {
		err := parseValue(rowJSON.Get(colName), colName, types[idx], row)
		if err != nil {
			return err
		}
	}
	return nil
}
