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

type Interceptor = func(*http.Request)

var beforeEachInterceptors []Interceptor

func BeforeEach(interceptor Interceptor) {
	beforeEachInterceptors = append(beforeEachInterceptors, interceptor)
}

type ContextKey struct{ Value string }

var CorrelationIDContextKey = &ContextKey{Value: "correlation_id_context_key"}

var CorrelationIDHeaderKey = "x-correlation-id"

var ErrUnexpectedResponseStatus = errors.New("expected response status to be in the 200-299 range")

type RequestBuilder struct {
	client   *http.Client
	ctx      context.Context
	method   string
	endpoint string
	request  *http.Request
	response *http.Response
	query    url.Values
	headers  http.Header
	body     io.Reader
	err      error
}

func (builder *RequestBuilder) Header(key, value string) *RequestBuilder {
	builder.headers[key] = append(builder.headers[key], value)

	return builder
}

func (builder *RequestBuilder) Query(key string, value interface{}) *RequestBuilder {
	switch value := value.(type) {
	case string:
		builder.query.Add(key, value)
	case []string:
		for _, v := range value {
			builder.query.Add(key, v)
		}
	}

	return builder
}

func (builder *RequestBuilder) Request() (*http.Request, error) {
	responseBuilder := builder.Build()

	if responseBuilder.err != nil {
		return responseBuilder.request, errors.WithStack(responseBuilder.err)
	}

	return responseBuilder.request, nil
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

	if correlationID, ok := builder.ctx.Value(CorrelationIDContextKey).(string); ok {
		req.Header.Add(CorrelationIDHeaderKey, correlationID)
	}

	req.URL.RawQuery = builder.query.Encode()

	out.request = req

	return out
}

func (builder *RequestBuilder) Send() (Response, error) {
	responseBuilder := builder.Build()

	responseBuilder.makeRequest()

	if responseBuilder.err != nil {
		return responseBuilder.response, errors.WithStack(responseBuilder.err)
	}

	return responseBuilder.response, nil
}

type ImpureRequestBuilder struct {
	RequestBuilder
}

func (builder *ImpureRequestBuilder) Body(reader io.Reader) *ImpureRequestBuilder {
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
	response Response
	err      error
}

type Response struct {
	*http.Response
	Body []byte
}

func (response *Response) JSON(out interface{}) error {
	if err := json.Unmarshal(response.Body, out); err != nil {
		return err
	}

	return nil
}

func (response *Response) Bytes() []byte {
	return response.Body
}

func (response *Response) Text() string {
	return string(response.Body)
}

func (builder *ResponseBuilder) makeRequest() {
	if builder.err != nil {
		return
	}

	for _, interceptor := range beforeEachInterceptors {
		interceptor(builder.request)
	}

	response, err := builder.client.Do(builder.request)

	builder.response = Response{Response: response}

	if err != nil {
		builder.err = errors.WithStack(err)
		return
	}

	body, _ := ioutil.ReadAll(response.Body)
	builder.response.Body = body

	response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode > 299 {
		builder.err = errors.Wrapf(ErrUnexpectedResponseStatus, "got status %d", response.StatusCode)
	}
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
