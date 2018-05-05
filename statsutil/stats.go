package statsutil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"path"
	"time"

	"h12.io/stats"
)

func Handler(s *stats.S, root string) http.Handler {
	mux := http.NewServeMux()
	mux.Handle(path.Join(root, "vars"), varsHandler(s))
	mux.Handle(path.Join(root, "pull"), pullHandler(s))
	return mux
}

func pullHandler(s *stats.S) http.Handler {
	client := http.Client{} // shared client
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		from := req.URL.Query().Get("from")
		resp, err := client.Get(from)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		otherStats := stats.New()
		if err := json.NewDecoder(resp.Body).Decode(otherStats); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		s.Merge(otherStats, time.Now().Add(-time.Duration(stats.DefaultBufferSize)*time.Second))
	})
}

func varsHandler(s *stats.S) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		buf, err := marshalJSON(s)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(buf)
	})
}

func marshalJSON(v interface{}) ([]byte, error) {
	b, err := json.Marshal(v)
	b = bytes.Replace(b, []byte(`\u003c`), []byte("<"), -1)
	b = bytes.Replace(b, []byte(`\u003e`), []byte(">"), -1)
	b = bytes.Replace(b, []byte(`\u0026`), []byte("&"), -1)
	return b, err
}
