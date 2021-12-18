package reqwest

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/pkg/errors"
)

type RequestBuilder struct {
	client   *http.Client
	ctx      context.Context
	method   string
	endpoint string
	request  *http.Request
	response *http.Response
	query    url.Values
	err      error
	headers  map[string][]string
}

func (builder *RequestBuilder) makeRequest() {
	if builder.err != nil {
		return
	}

	response, err := builder.client.Do(builder.request)

	builder.err = err
	builder.response = response
}

func (builder *RequestBuilder) Header(key, value string) *RequestBuilder {
	if builder.headers == nil {
		builder.headers = make(map[string][]string)
	}

	builder.headers[key] = append(builder.headers[key], value)

	return builder
}

func (builder *RequestBuilder) Build() *ResponseBuilder {
	out := &ResponseBuilder{requestBuilder: builder}

	if builder.err != nil {
		out.requestBuilder.err = errors.WithStack(builder.err)
		return out
	}

	req, err := http.NewRequest(builder.method, builder.endpoint, nil)
	if err != nil {
		out.requestBuilder.err = errors.WithStack(err)
		return out
	}

	req = req.WithContext(builder.ctx)

	req.Header = builder.headers

	return &ResponseBuilder{requestBuilder: builder, request: req}
}

type ImpureRequestBuilder struct {
	RequestBuilder
}

func (builder *ImpureRequestBuilder) Header(key, value string) *ImpureRequestBuilder {
	builder.RequestBuilder.Header(key, value)

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
	request        *http.Request
}

func (builder *ResponseBuilder) Request() *http.Request {
	return builder.request
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
			ctx:     ctx,
			request: req,
			client:  http.DefaultClient,
			method:  http.MethodPost,
			err:     err,
		},
	}
}

type PureRequestBuilder struct {
	RequestBuilder
}

func (builder *PureRequestBuilder) Header(key, value string) *PureRequestBuilder {
	builder.RequestBuilder.Header(key, value)
	return builder
}

func GET(ctx context.Context, endpoint string) *PureRequestBuilder {

	return &PureRequestBuilder{
		RequestBuilder: RequestBuilder{
			ctx:      ctx,
			endpoint: endpoint,
			client:   http.DefaultClient,
			method:   http.MethodGet,
			query:    url.Values{},
		},
	}
}
