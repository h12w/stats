package stats

import (
	"net/url"
	"strings"
)

type (
	Key  string
	Tags map[string]string
)

func NewKey(name string, tags Tags) Key {
	key := Key(name)
	if len(tags) > 0 {
		key += " " + Key(tags.encode())
	}
	return key
}

func (key Key) Decode() (name string, tags Tags, err error) {
	parts := strings.SplitN(string(key), " ", 2)
	name = parts[0]
	if len(parts) == 1 {
		return name, Tags{}, nil
	}
	tags, err = parseTags(parts[1])
	if err != nil {
		return name, nil, err
	}
	return name, tags, nil
}

func (tags Tags) encode() Key {
	values := make(url.Values)
	for key, value := range tags {
		values.Set(key, value)
	}
	return Key(values.Encode())
}

func parseTags(s string) (Tags, error) {
	values, err := url.ParseQuery(s)
	if err != nil {
		return nil, err
	}
	tags := make(Tags)
	for key := range values {
		tags[key] = values.Get(key)
	}
	return tags, nil
}
