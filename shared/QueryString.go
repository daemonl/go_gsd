package shared

import (
	"net/url"
	"regexp"
	"strconv"
	"time"
)

type IQueryString interface {
	Int64(string) (int64, bool)
	Timestamp(string) (time.Time, bool)
	Date(string) (time.Time, bool)
	String(string) (string, bool)
}

func GetQueryString(values url.Values) IQueryString {
	return &queryString{
		values: values,
	}
}

type queryString struct {
	values url.Values
}

// String gets a string from the query values
func (qs *queryString) String(key string) (string, bool) {
	str := qs.values.Get(key)
	if len(str) < 1 {
		return "", false
	}
	return str, true
}

// Int64 parses an int64 from the querystring, if it exists and is parsable
func (qs *queryString) Int64(key string) (int64, bool) {
	str, ok := qs.String(key)
	if !ok {
		return 0, false
	}
	val, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, false
	}
	return val, true
}

// Timestamp parses a timestamp from the querystring as a time.Time, if it exists and can be parsed
func (qs *queryString) Timestamp(key string) (time.Time, bool) {
	i, ok := qs.Int64(key)
	if !ok {
		return time.Time{}, false
	}
	return time.Unix(i, 0), true
}

var reNotNumber *regexp.Regexp = regexp.MustCompile(`[^0-9]`)

// Date parses a date in yyyy-mm-dd format, it strips any non numerics, so the separators don't matter at all
// i.e. "2006-01-02" "20060102" and "2006a0b1c0d2e" are all the same
func (qs *queryString) Date(key string) (time.Time, bool) {
	str, ok := qs.String(key)
	if !ok {
		return time.Time{}, false
	}
	str = reNotNumber.ReplaceAllString(str, "")
	t, err := time.Parse("20060102", str)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}
