package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

type RequestBuilder struct {
	client   *http.Client
	method   string
	request  *http.Request
	response *http.Response
	query    url.Values
	err      error
}

func (builder *RequestBuilder) makeRequest() {
	if builder.err != nil {
		return
	}

	response, err := builder.client.Do(builder.request)

	builder.err = err
	builder.response = response
}

func (builder *RequestBuilder) Send() *ResponseBuilder {
	return &ResponseBuilder{requestBuilder: builder}
}

type ImpureRequestBuilder struct {
	RequestBuilder
}

func (builder *ImpureRequestBuilder) Header(key, value string) *ImpureRequestBuilder {
	builder.request.Header.Set(key, value)
	return builder
}

func (builder *ImpureRequestBuilder) JSON(body interface{}) *ImpureRequestBuilder {
	if builder.err != nil {
		return builder
	}

	builder.Header("content-type", "application/json")

	builder.makeRequest()

	if builder.err != nil {
		return builder
	}

	out, err := json.Marshal(body)
	if err != nil {
		builder.err = err
		return builder
	}

	builder.request.Body = io.NopCloser(bytes.NewBuffer(out))

	return builder
}

type ResponseBuilder struct {
	requestBuilder *RequestBuilder
}

func (builder *ResponseBuilder) JSON(out interface{}) error {
	builder.requestBuilder.request.Header.Set("accept", "application/json")

	builder.requestBuilder.makeRequest()

	if builder.requestBuilder.err != nil {
		return builder.requestBuilder.err
	}

	defer builder.requestBuilder.response.Body.Close()

	body, err := ioutil.ReadAll(builder.requestBuilder.response.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, out); err != nil {
		return err
	}

	return nil
}

func POST(ctx context.Context, endpoint string) *ImpureRequestBuilder {
	req, err := http.NewRequest(http.MethodPost, endpoint, nil)

	req = req.WithContext(ctx)

	return &ImpureRequestBuilder{
		RequestBuilder: RequestBuilder{
			request: req,
			client:  http.DefaultClient,
			method:  http.MethodPost,
			query:   url.Values{},
			err:     err,
		},
	}
}

type PureRequestBuilder struct {
	RequestBuilder
}

func (builder *PureRequestBuilder) Header(key, value string) *PureRequestBuilder {
	builder.request.Header.Set(key, value)
	return builder
}

func GET(ctx context.Context, endpoint string) *PureRequestBuilder {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)

	req = req.WithContext(ctx)

	return &PureRequestBuilder{
		RequestBuilder: RequestBuilder{
			request: req,
			client:  http.DefaultClient,
			method:  http.MethodGet,
			query:   url.Values{},
			err:     err,
		},
	}
}
