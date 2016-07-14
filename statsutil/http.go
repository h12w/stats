package statsutil

import (
	"encoding/json"
	"net/http"

	"h12.me/stats"
)

func Handler(s *stats.S) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		buf, err := json.Marshal(s)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(buf)
	})
}
