package statsutil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
	"h12.me/stats"
)

type Host struct {
	URLs []string
	Tag  string
}

func CollectStats(httpClient *http.Client, hosts []Host, start time.Time) (*stats.S, error) {
	allStats := stats.New()
	var g errgroup.Group
	for i := range hosts {
		host := &hosts[i]
		g.Go(func() error {
			s, err := host.get(httpClient)
			if err != nil {
				return err
			}
			return allStats.MergeWithTags(s, start, stats.Tags{"host": host.Tag})
		})
	}
	return allStats, g.Wait()
}

func (h *Host) get(client *http.Client) (*stats.S, error) {
	var (
		resp *http.Response
		err  error
	)
	for _, uri := range h.URLs {
		resp, err = client.Get(uri)
		if err != nil {
			continue
		}
		break
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %v", err.Error(), h.URLs)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%d: %v", resp.StatusCode, h.URLs)
	}
	s := stats.New()
	return s, json.NewDecoder(resp.Body).Decode(s)
}
