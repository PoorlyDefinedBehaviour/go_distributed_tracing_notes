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
	headers  http.Header
	body     io.Reader
}

func (builder *RequestBuilder) Header(key, value string) *RequestBuilder {
	builder.headers[key] = append(builder.headers[key], value)

	return builder
}

func (builder *RequestBuilder) Build() *ResponseBuilder {
	out := &ResponseBuilder{
		client: builder.client,
	}

	if builder.err != nil {
		out.err = errors.WithStack(builder.err)
		return out
	}

	req, err := http.NewRequest(builder.method, builder.endpoint, builder.body)
	if err != nil {
		out.err = errors.WithStack(err)
		return out
	}

	req = req.WithContext(builder.ctx)

	if len(builder.headers) > 0 {
		req.Header = builder.headers
	}

	out.request = req

	return out
}

type ImpureRequestBuilder struct {
	RequestBuilder
}

func (builder *ImpureRequestBuilder) Body(reader io.Reader) *ImpureRequestBuilder {
	if builder.err != nil {
		return builder
	}

	builder.body = reader

	return builder
}

func (builder *ImpureRequestBuilder) Header(key, value string) *ImpureRequestBuilder {
	builder.RequestBuilder.Header(key, value)

	return builder
}

func (builder *ImpureRequestBuilder) JSON(body interface{}) *ImpureRequestBuilder {
	if builder.err != nil {
		return builder
	}

	out, err := json.Marshal(body)
	if err != nil {
		builder.err = errors.WithStack(err)
		return builder
	}

	builder.Body(bytes.NewBuffer(out))

	return builder
}

type ResponseBuilder struct {
	client   *http.Client
	request  *http.Request
	response *http.Response
	err      error
}

func (builder *ResponseBuilder) makeRequest() {
	if builder.err != nil {
		return
	}

	response, err := builder.client.Do(builder.request)
	if err != nil {
		builder.err = errors.WithStack(err)
	}

	builder.response = response
}

func (builder *ResponseBuilder) Request() *http.Request {
	return builder.request
}

func (builder *ResponseBuilder) Response() (*http.Response, error) {
	builder.makeRequest()

	if builder.err != nil {
		return builder.response, errors.WithStack(builder.err)
	}

	return builder.response, nil
}

func (builder *ResponseBuilder) Bytes() ([]byte, error) {
	builder.makeRequest()

	var out []byte

	if builder.err != nil {
		return out, errors.WithStack(builder.err)
	}

	defer builder.response.Body.Close()

	out, err := ioutil.ReadAll(builder.response.Body)
	if err != nil {
		builder.err = errors.WithStack(err)
		return out, builder.err
	}

	return out, nil
}

func (builder *ResponseBuilder) Text() (string, error) {
	out, err := builder.Bytes()
	if err != nil {
		builder.err = errors.WithStack(err)
		return string(out), builder.err
	}

	return string(out), nil
}

func (builder *ResponseBuilder) JSON(out interface{}) error {
	if builder.err != nil {
		return errors.WithStack(builder.err)
	}

	builder.request.Header.Set("accept", "application/json")

	builder.makeRequest()

	if builder.err != nil {
		return builder.err
	}

	defer builder.response.Body.Close()

	body, err := ioutil.ReadAll(builder.response.Body)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, out); err != nil {
		return err
	}

	return nil
}

func POST(ctx context.Context, endpoint string) *ImpureRequestBuilder {
	return &ImpureRequestBuilder{
		RequestBuilder: RequestBuilder{
			ctx:      ctx,
			endpoint: endpoint,
			client:   http.DefaultClient,
			method:   http.MethodPost,
			query:    url.Values{},
			headers:  make(http.Header, 0),
		},
	}
}

func PATCH(ctx context.Context, endpoint string) *ImpureRequestBuilder {
	return &ImpureRequestBuilder{
		RequestBuilder: RequestBuilder{
			ctx:      ctx,
			endpoint: endpoint,
			client:   http.DefaultClient,
			method:   http.MethodPatch,
			headers:  make(http.Header, 0),
		},
	}
}

func PUT(ctx context.Context, endpoint string) *ImpureRequestBuilder {
	return &ImpureRequestBuilder{
		RequestBuilder: RequestBuilder{
			ctx:      ctx,
			endpoint: endpoint,
			client:   http.DefaultClient,
			method:   http.MethodPut,
			headers:  make(http.Header, 0),
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
			headers:  make(http.Header, 0),
		},
	}
}
