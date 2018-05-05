package influx

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"h12.io/stats"
)

type DB struct {
	c        client.Client
	database string
	bp       client.BatchPoints
	mu       sync.Mutex
}

// Open oepns an influxdb client with a connection string like:
// (http|udp)://[username:password@]host/path[?args], and args can be any of:
// timeout=duration_string, ua=user_agent
func Open(dataSourceName string) (*DB, error) {
	uri, err := url.Parse(dataSourceName)
	if err != nil {
		return nil, err
	}
	switch uri.Scheme {
	case "http":
		var (
			user     string
			password string
		)
		if uri.User != nil {
			user = uri.User.Username()
			password, _ = uri.User.Password()
		}
		query := uri.Query()
		timeout, _ := time.ParseDuration(query.Get("timeout"))
		addr := "http://" + uri.Host
		c, err := client.NewHTTPClient(client.HTTPConfig{
			Addr:      addr,
			Username:  user,
			Password:  password,
			UserAgent: query.Get("ua"),
			Timeout:   timeout,
		})
		if err != nil {
			return nil, err
		}
		database := strings.TrimPrefix(uri.Path, `/`)
		resp, err := c.Query(client.NewQuery("CREATE DATABASE "+database, "", ""))
		if err != nil {
			return nil, err
		}
		if resp.Error() != nil {
			return nil, resp.Error()
		}
		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Precision: "ns",
			Database:  database,
		})
		if err != nil {
			return nil, err
		}
		return &DB{
			c:        c,
			database: database,
			bp:       bp,
		}, nil
	}
	// TODO: support https, udp
	return nil, fmt.Errorf("unsupported scheme %s", uri.Scheme)
}

func (d *DB) Close() error {
	return d.c.Close()
}

func (d *DB) SaveStats(s *stats.S, from time.Time, du time.Duration, tags map[string]string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	start := int(from.Unix())
	end := int(from.Add(du).Unix())
	for key, meter := range s.Meters {
		name, meterTags, err := key.Decode()
		if err != nil {
			return err
		}
		for key, val := range tags {
			meterTags[key] = val
		}
		for sec := start; sec < end; sec++ {
			measure := meter.Get(sec)
			if measure == 0 {
				continue
			}
			d.insert(name, time.Unix(int64(sec), 0), meterTags, map[string]interface{}{"count": measure})
		}
	}
	return d.commit()
}

func (d *DB) insert(tableName string, t time.Time, tags map[string]string, fields map[string]interface{}) error {
	p, err := client.NewPoint(tableName, tags, fields, t)
	if err != nil {
		return err
	}
	d.bp.AddPoint(p)
	return nil
}

func (d *DB) commit() error {
	if err := d.c.Write(d.bp); err != nil {
		return err
	}
	var err error
	d.bp, err = client.NewBatchPoints(client.BatchPointsConfig{
		Precision: "ns",
		Database:  d.database,
	})
	return err
}
