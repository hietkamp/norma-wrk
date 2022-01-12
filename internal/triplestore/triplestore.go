package triplestore

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Options struct {
	BaseURL string
}

type Triplestore struct {
	HttpClient *http.Client
	BaseUrl    string
}

func New(options Options) *Triplestore {
	url, err := url.ParseRequestURI(options.BaseURL)
	if err != nil {
		panic(fmt.Errorf("failed on parsing base url: %w", err))
	}
	return &Triplestore{
		HttpClient: &http.Client{},
		BaseUrl:    url.String(),
	}
}

func (s *Triplestore) RunQuery(query string) (interface{}, error) {
	var jsonResultSet interface{}

	urlValues := make(url.Values)
	urlValues.Add("query", query)
	req, err := http.NewRequest("GET", fmt.Sprintf("%s?%s", s.BaseUrl, urlValues.Encode()), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/sparql-results+json")
	resp, err := s.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to run query: service not available")
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(body, &jsonResultSet)
	return jsonResultSet, nil
}
