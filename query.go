package cloudtrailquery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail/types"
)

type options struct {
	normalize                  bool
	expand                     bool
	queryRunningFrequencyCheck time.Duration
}

var defaults = options{
	normalize:                  true,
	expand:                     true,
	queryRunningFrequencyCheck: 250 * time.Millisecond,
}

type Option = func(*options)

func Normalize(normalize bool) Option {
	return func(o *options) {
		o.normalize = normalize
	}
}

func Expand(expand bool) Option {
	return func(o *options) {
		o.expand = expand
	}
}

func RunningFrequencyCheck(duration time.Duration) Option {
	return func(o *options) {
		o.queryRunningFrequencyCheck = duration
	}
}

type Row = map[string]any

func QueryAll(ctx context.Context, client *cloudtrail.Client, query string, opts ...Option) ([]Row, error) {
	results := []Row{}
	callback := func(row Row) error {
		results = append(results, row)
		return nil
	}
	err := Query(ctx, client, query, callback, opts...)
	return results, err
}

func QueryStream(ctx context.Context, client *cloudtrail.Client, query string, writer io.Writer, opts ...Option) error {
	encoder := json.NewEncoder(writer)
	callback := func(row Row) error {
		if err := encoder.Encode(row); err != nil {
			return err
		}
		return nil
	}
	return Query(ctx, client, query, callback, opts...)
}

func Query(ctx context.Context, client *cloudtrail.Client, query string, callback func(Row) error, opts ...Option) error {
	if client == nil {
		return errors.New("nil cloudtrail client")
	}
	if callback == nil {
		return errors.New("nil callback")
	}

	o := defaults
	for _, opt := range opts {
		opt(&o)
	}

	log.Printf("running query: %s", query)
	runningQuery, err := client.StartQuery(ctx, &cloudtrail.StartQueryInput{
		QueryStatement: &query,
	})
	if err != nil {
		return err
	}

	queryID := runningQuery.QueryId
	log.Printf("query id: %s", *queryID)

	var pageToken *string
	var maxResults int32 = 1000

	var processed int32
	var total int32

	for {
		start := time.Now()
		results, err := client.GetQueryResults(ctx, &cloudtrail.GetQueryResultsInput{
			QueryId:         queryID,
			MaxQueryResults: &maxResults,
			NextToken:       pageToken,
		})
		if err != nil {
			return err
		}

		switch results.QueryStatus {
		case types.QueryStatusCancelled:
			return errors.New("query cancelled")
		case types.QueryStatusFailed:
			return fmt.Errorf("query failed: %w", errors.New(*results.ErrorMessage))
		case types.QueryStatusTimedOut:
			return errors.New("query timed out")
		}

		for _, raw := range results.QueryResultRows {
			row := map[string]any{}
			for _, entry := range raw {
				for k, v := range entry {
					row[k] = v
				}
			}
			if o.expand {
				expandRow(row)
			}
			if o.normalize {
				normalizeJson(row)
			}

			if err := callback(row); err != nil {
				return err
			}

			processed++
		}

		if total == 0 && results.QueryStatistics != nil && results.QueryStatistics.TotalResultsCount != nil {
			total = *results.QueryStatistics.TotalResultsCount
		}
		if processed > 0 {
			if total != 0 {
				log.Printf("progress: %d/%d (%.2f%%)", processed, total, float64(processed)/float64(total)*100)
			} else {
				log.Printf("progress: %d/? (?%%)", processed)
			}
		}

		pageToken = results.NextToken
		if pageToken == nil {
			break
		}

		// To avoid issuing too many AWS api calls to get query results while the query is being
		// processed on AWS side
		if len(results.QueryResultRows) == 0 {
			select {
			case <-time.After(o.queryRunningFrequencyCheck - time.Since(start)):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return nil
}

var jsonLikeKeyPattern = regexp.MustCompile(`(\w+)=`)
var jsonLikeValuePattern = regexp.MustCompile(`=([^{][^,}]+)`)
var jsonLikeSepPattern = regexp.MustCompile(`"=(["[{])`)

func expandRow(row map[string]any) {
	set := map[string]any{}
	for key, value := range row {
		str, ok := value.(string)
		if !ok || str == "" {
			continue
		}
		newStr := jsonLikeKeyPattern.ReplaceAllString(str, "\"$1\"=")
		newStr = jsonLikeValuePattern.ReplaceAllString(newStr, "=\"$1\"")
		if newStr == str {
			continue
		}
		newStr = jsonLikeSepPattern.ReplaceAllString(newStr, "\":$1")

		var newValue any
		if err := json.Unmarshal([]byte(newStr), &newValue); err != nil {
			continue
		}
		set[key+"__parsed"] = newValue
	}
	for key, value := range set {
		row[key] = value
	}
}

func normalizeJson(value any) any {
	if m, ok := value.(map[string]any); ok {
		for k, v := range m {
			m[k] = normalizeJson(v)
		}
		return m
	} else if s, ok := value.([]any); ok {
		for i, v := range s {
			s[i] = normalizeJson(v)
		}
		return s
	} else if str, ok := value.(string); ok {
		switch str {
		case "":
			return nil
		case "null":
			return nil
		case "true":
			return true
		case "false":
			return false
		}
	}
	return value
}
