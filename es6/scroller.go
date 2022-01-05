package es6

import (
	"bytes"
	"context"
	"fmt"

	es6api "github.com/elastic/go-elasticsearch/v6/esapi"
	sifes "github.com/go-sif/sif-datasource-elasticsearch"
	"github.com/tidwall/gjson"
)

type es6scroller struct {
	client    *es6client
	queryCopy *es6api.SearchRequest
	conf      *sifes.DataSourceConf
	shard     int64
	scrollID  string
	finished  bool
}

func (s *es6scroller) IsFinished() bool {
	return s.finished
}

func (s *es6scroller) ScrollDocuments(ctx context.Context) ([]gjson.Result, error) {
	if s.finished {
		return nil, fmt.Errorf("finished")
	}

	// if we're already scrolling, then continue
	if s.scrollID != "" {
		// otherwise, scroll next document
		res, err := s.client.api.Scroll(
			s.client.api.Scroll.WithScrollID(s.scrollID),
			s.client.api.Scroll.WithScroll(s.conf.ScrollTimeout),
		)
		if err != nil {
			return nil, fmt.Errorf("Unable to scroll documents: %s", err)
		}
		if res.IsError() {
			return nil, fmt.Errorf("Unable to scroll documents: %s", res)
		}
		return s.parseScrollResponse(res)
	}

	// otherwise this is the first call to scroll
	// and we need to initiate a scroll by issuing the query
	s.queryCopy.Preference = fmt.Sprintf("_shards:%d", s.shard)
	s.queryCopy.Size = &s.conf.PartitionSize // TODO set a maximum size?
	s.queryCopy.Scroll = s.conf.ScrollTimeout
	res, err := s.queryCopy.Do(ctx, s.client.api)
	if err != nil {
		return nil, err
	}
	if res.IsError() {
		return nil, fmt.Errorf("Unable to scroll documents: %s", res)
	}
	return s.parseScrollResponse(res)
}

func (s *es6scroller) parseScrollResponse(res *es6api.Response) ([]gjson.Result, error) {
	defer res.Body.Close()
	var b bytes.Buffer
	_, err := b.ReadFrom(res.Body)
	if err != nil {
		return nil, err
	}
	// parse response body
	body := b.String()
	// store scroll ID for next call
	s.scrollID = gjson.Get(body, "_scroll_id").String()
	// check number of results
	hits := gjson.Get(body, "hits.hits").Array()
	if len(hits) < 1 {
		// close scroll
		s.client.api.ClearScroll(s.client.api.ClearScroll.WithScrollID(s.scrollID))
		s.scrollID = ""
		s.finished = true
		// return
		return nil, nil
	}
	documents := gjson.Get(body, "hits.hits").Array()
	return documents, nil
}
