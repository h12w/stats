package statsutil

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"

	"h12.me/stats"
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
		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(buf))
		if err := json.Unmarshal(buf, otherStats); err != nil {
			log.Fatal(err)
		}
		/*
			if err := json.NewDecoder(resp.Body).Decode(otherStats); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		*/
		fmt.Println(otherStats)
		s.Merge(otherStats)
		fmt.Println(s)
	})
}

func varsHandler(s *stats.S) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		buf, err := json.Marshal(s)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(buf)
	})
}
