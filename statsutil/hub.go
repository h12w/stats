package statsutil

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/sync/errgroup"
	"h12.me/stats"
)

type Host struct {
	URL string `json:"url"`
	Tag string `json:"tag"`
}

func CollectStats(httpClient *http.Client, hosts []Host) (*stats.S, error) {
	allStats := stats.New()
	var g errgroup.Group
	for i := range hosts {
		host := &hosts[i]
		g.Go(func() error {
			s, err := host.get(httpClient)
			if err != nil {
				return err
			}
			allStats.MergeWithTags(s, stats.Tags{"host": host.Tag})
			return nil
		})
	}
	return allStats, g.Wait()
}
func (h *Host) get(client *http.Client) (*stats.S, error) {
	resp, err := client.Get(h.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%d: %s", resp.StatusCode, h.URL)
	}
	s := stats.New()
	return s, json.NewDecoder(resp.Body).Decode(s)
}
