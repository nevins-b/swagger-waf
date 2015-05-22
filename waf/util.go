package waf

import (
	"errors"
	"strconv"
	"time"
)

func stringInSlice(slice []string, s string) bool {
	for i := range slice {
		if slice[i] == s {
			return true
		}
	}
	return false
}

func stringToType(in, format string) (interface{}, error) {
	switch format {
	case "int32":
		i, err := strconv.ParseInt(in, 10, 32)
		return int32(i), err
	case "int64":
		return strconv.ParseInt(in, 10, 64)
	case "float":
		return strconv.ParseFloat(in, 64)
	case "double":
		return strconv.ParseFloat(in, 64)
	case "byte":
		return []byte(in), nil
	case "date":
		return time.Parse(time.RFC3339, in)
	case "date-time":
		return time.Parse(time.RFC3339, in)
	default:
		return nil, errors.New("Unable to convert")
	}
}
