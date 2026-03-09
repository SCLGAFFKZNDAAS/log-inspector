package loki

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

var httpClient = &http.Client{}

func QueryLoki(query LokiQuery) (LokiResponse, error) {
	LOKI_URL := os.Getenv("LOKI_URL")
	req, err := http.NewRequest("GET", LOKI_URL+"/loki/api/v1/query_range", nil)
	if err != nil {
		return LokiResponse{}, err
	}

	q := req.URL.Query()
	q.Add("query", query.Query)
	q.Add("start", fmt.Sprintf("%d", query.Start.UnixNano()))
	q.Add("end", fmt.Sprintf("%d", query.End.UnixNano()))
	q.Add("limit", fmt.Sprintf("%d", query.Limit))
	req.URL.RawQuery = q.Encode()

	resp, err := httpClient.Do(req)
	if err != nil {
		return LokiResponse{}, err
	}
	defer resp.Body.Close()

	fmt.Println("Response status:", resp.Status)

	var lokiResp LokiResponse
	if err := json.NewDecoder(resp.Body).Decode(&lokiResp); err != nil {
		return LokiResponse{}, err
	}

	return lokiResp, nil
}
