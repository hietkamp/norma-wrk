package handlers

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hietkamp/norma-wrk/internal/eventstream"
	"github.com/hietkamp/norma-wrk/internal/triplestore"
	"github.com/rs/zerolog/log"
)

func RunQuery(v ValidatedQueriesReceived) (QueryResult, error) {
	// The sparql is url encoded. Unescape it, because the sparql adapter will also encode it.
	sparql, err := url.QueryUnescape(v.Payload.CredentialSubject.ValidatetQuery.Sparql)
	if err != nil {
		return QueryResult{}, fmt.Errorf("failed to unescape query: %w", err)
	}
	log.Debug().Msgf("Sparql reveived: %s", string(sparql))
	log.Debug().Msgf("Run the sparql on endpoint %s", os.Getenv("sparql_endpoint"))
	sparqlServer := triplestore.New(triplestore.Options{BaseURL: os.Getenv("sparql_endpoint")})
	result, err := sparqlServer.RunQuery(sparql)
	if err != nil {
		return QueryResult{}, fmt.Errorf("failed to run query: %w", err)
	}
	return QueryResult{Result: result}, nil
}

func ProduceMessage(v ValidatedQueriesProcessed) error {
	messageBytes, _ := json.Marshal(v)
	brokers := strings.Split(os.Getenv("kafka_url"), ",")
	es := eventstream.New("tcp", brokers[1])
	err := es.Produce(os.Getenv("topic_producer"), messageBytes)
	if err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}
	return nil
}

func HandleValidatedQueryReceived(message []byte) error {
	var vqReceived ValidatedQueriesReceived
	json.Unmarshal(message, &vqReceived)

	// Run the query on the triplestore
	result, err := RunQuery(vqReceived)
	if err != nil {
		return err
	}
	// Compose the resultset
	ts := time.Now()
	vqProcessed := ValidatedQueriesProcessed{
		Timestamp: ts.Format(time.RFC3339),
		Header:    vqReceived.Header,
	}
	vqProcessed.Payload = append(vqProcessed.Payload, result)
	err = ProduceMessage(vqProcessed)
	if err != nil {
		return err
	}
	return nil
}
