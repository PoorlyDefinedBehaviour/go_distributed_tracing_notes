package reqwest

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/IQ-tech/go-datagen"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func randomEndpoint() string {
	return fmt.Sprintf("http://localhost:5000/%s/", datagen.StringWithAlphabetic(10))
}

func Test_BeforeEach(t *testing.T) {
	t.Parallel()

	t.Run("does not call interceptors if builder is in error state", func(t *testing.T) {
		BeforeEach(func(_ *http.Request) {
			panic("called")
		})

		responseBuilder := GET(context.Background(), randomEndpoint())
		responseBuilder.err = errors.New("some error")

		_, _ = responseBuilder.Send()

		beforeEachInterceptors = make([]Interceptor, 0)
	})

	t.Run("passes request to interceptors before request is sent", func(t *testing.T) {
		defer gock.Off()

		endpoints := []string{
			"http://localhost:5000/test_before_each_1",
			"http://localhost:5000/test_before_each_2",
			"http://localhost:5000/test_before_each_3",
		}

		endpointsCalled := make([]string, 0, len(endpoints))

		BeforeEach(func(req *http.Request) {
			endpointsCalled = append(endpointsCalled, req.URL.String())
		})

		for _, endpoint := range endpoints {
			gock.New(endpoint).
				Get("").
				Reply(200).
				Body(strings.NewReader("hello world"))

			_, _ = GET(context.Background(), endpoint).Send()
		}

		assert.Equal(t, endpoints, endpointsCalled)
	})
}

func Test_CreatesRequestBuilder(t *testing.T) {
	t.Parallel()

	t.Run("GET", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		endpoint := randomEndpoint()

		builder := GET(ctx, endpoint)

		assert.Equal(t, ctx, builder.ctx)
		assert.Equal(t, endpoint, builder.endpoint)
		assert.Equal(t, http.DefaultClient, builder.client)
		assert.Equal(t, http.MethodGet, builder.method)
		assert.Empty(t, builder.query)
		assert.NotNil(t, builder.headers)
		assert.Nil(t, builder.err)
	})

	t.Run("POST", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		endpoint := randomEndpoint()

		builder := POST(ctx, endpoint)

		assert.Equal(t, ctx, builder.ctx)
		assert.Equal(t, endpoint, builder.endpoint)
		assert.Equal(t, http.DefaultClient, builder.client)
		assert.Equal(t, http.MethodPost, builder.method)
		assert.Empty(t, builder.query)
		assert.NotNil(t, builder.headers)
		assert.Nil(t, builder.err)
	})

	t.Run("PATCH", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		endpoint := randomEndpoint()

		builder := PATCH(ctx, endpoint)

		assert.Equal(t, ctx, builder.ctx)
		assert.Equal(t, endpoint, builder.endpoint)
		assert.Equal(t, http.DefaultClient, builder.client)
		assert.Equal(t, http.MethodPatch, builder.method)
		assert.Empty(t, builder.query)
		assert.NotNil(t, builder.headers)
		assert.Nil(t, builder.err)
	})

	t.Run("PUT", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		endpoint := randomEndpoint()

		builder := PUT(ctx, endpoint)

		assert.Equal(t, ctx, builder.ctx)
		assert.Equal(t, endpoint, builder.endpoint)
		assert.Equal(t, http.DefaultClient, builder.client)
		assert.Equal(t, http.MethodPut, builder.method)
		assert.Empty(t, builder.query)
		assert.NotNil(t, builder.headers)
		assert.Nil(t, builder.err)
	})
}

func Test_PureRequestBuilder_Header(t *testing.T) {
	t.Parallel()

	t.Run("adds header to request", func(t *testing.T) {
		request, err := GET(context.Background(), "https://api.github.com/users/poorlydefinedbehaviour/repos").
			Header("key1", "value1").
			Header("key2", "value2").
			Header("key3", "VALUE3").
			Request()

		expected := map[string][]string{
			"key1": {"value1"},
			"key2": {"value2"},
			"key3": {"VALUE3"},
		}

		assert.NoError(t, err)
		assert.EqualValues(t, expected, request.Header)
	})
}

func Test_PureRequestBuilder_Query(t *testing.T) {
	t.Parallel()

	t.Run("adds query string to request url", func(t *testing.T) {
		request, err := GET(context.Background(), "https://api.github.com/users/poorlydefinedbehaviour/repos").
			Query("key1", "value1").
			Query("key2", []string{"value1", "value2"}).
			Request()

		expected := url.Values{
			"key1": {"value1"},
			"key2": {"value1", "value2"},
		}

		assert.NoError(t, err)
		assert.Equal(t, expected.Encode(), request.URL.RawQuery)
	})
}

func Test_PureRequestBuilder_Body(t *testing.T) {
	t.Parallel()

	t.Run("adds any io.Reader to request body", func(t *testing.T) {
		t.Parallel()

		payload := "hello world"

		request, err := POST(context.Background(), "https://api.github.com/users/poorlydefinedbehaviour/repos").
			Body(strings.NewReader(payload)).
			Request()

		assert.NoError(t, err)

		requestBody, err := ioutil.ReadAll(request.Body)

		assert.NoError(t, err)

		assert.EqualValues(t, payload, requestBody)
	})
}

func Test_ImpureRequestBuilder_Header(t *testing.T) {
	t.Parallel()

	t.Run("adds header to request", func(t *testing.T) {
		request, err := POST(context.Background(), "https://api.github.com/users/poorlydefinedbehaviour/repos").
			Header("key1", "value1").
			Header("key2", "value2").
			Header("key3", "VALUE3").
			Request()

		expected := map[string][]string{
			"key1": {"value1"},
			"key2": {"value2"},
			"key3": {"VALUE3"},
		}

		fmt.Printf("\n\naaaaaaa request %+v\n\n", request)

		assert.NoError(t, err)
		assert.EqualValues(t, expected, request.Header)
	})
}

func Test_ImpureRequestBuilder_JSON(t *testing.T) {
	t.Parallel()

	t.Run("marshals json and adds it to request body", func(t *testing.T) {
		t.Parallel()

		payload := map[string]string{
			"hello": "world",
		}

		request, err := POST(context.Background(), "https://api.github.com/users/poorlydefinedbehaviour/repos").
			JSON(payload).
			Request()

		assert.NoError(t, err)

		requestBody, err := ioutil.ReadAll(request.Body)

		assert.NoError(t, err)

		out := make(map[string]string)

		json.Unmarshal(requestBody, &out)

		assert.EqualValues(t, payload, out)
	})
}

func Test_ResponseBuilder_makeRequest(t *testing.T) {
	t.Parallel()

	t.Run("when response status is not in the 200-299 range", func(t *testing.T) {
		t.Parallel()

		t.Run("returns custom error", func(t *testing.T) {
			t.Parallel()

			defer gock.Off()

			for _, status := range []int{103, 300} {
				endpoint := randomEndpoint()

				gock.New(endpoint).
					Get("").
					Reply(status)

				response, err := GET(context.Background(), endpoint).Send()

				assert.True(t, errors.Is(err, ErrUnexpectedResponseStatus))
				assert.Equal(t, status, response.StatusCode)
			}
		})

		t.Run("consumes response body and closes it", func(t *testing.T) {
			t.Parallel()
		})
	})
}

func Test_ResponseBuilder_Build(t *testing.T) {
	t.Parallel()

	t.Run("if context has a correlation id, adds it to the request", func(t *testing.T) {
		t.Parallel()

		requestID := "03823a30-5bf8-4cd9-ac53-d12d18ab6d3d"

		ctx := context.WithValue(context.Background(), CorrelationIDContextKey, requestID)

		request, err := GET(ctx, randomEndpoint()).
			Request()

		assert.NoError(t, err)
		assert.Equal(t, request.Header.Get(CorrelationIDHeaderKey), requestID)
	})
}

func Test_ResponseBuilder_Request(t *testing.T) {
	t.Parallel()

	t.Run("returns the http request", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		endpoint := randomEndpoint()

		request, err := GET(ctx, endpoint).Request()

		assert.NoError(t, err)
		assert.Equal(t, ctx, request.Context())
		assert.Equal(t, http.MethodGet, request.Method)
		assert.Equal(t, endpoint, request.URL.String())
	})

	t.Run("returns error if an error happened in the process", func(t *testing.T) {
		t.Parallel()

		builder := GET(context.Background(), randomEndpoint())

		expectedErr := errors.New("some error")

		builder.err = expectedErr

		_, err := builder.Request()

		assert.True(t, errors.Is(err, expectedErr))
	})
}

func Test_Response_Text(t *testing.T) {
	t.Parallel()

	t.Run("returns response body as text", func(t *testing.T) {
		t.Parallel()

		defer gock.Off()

		endpoint := randomEndpoint()

		gock.New(endpoint).
			Get("").
			Reply(200).
			JSON(map[string]string{"foo": "bar"})

		response, err := GET(context.Background(), endpoint).Send()

		assert.NoError(t, err)

		assert.JSONEq(t, `{"foo":"bar"}`, response.Text())
	})
}

func Test_Response_Bytes(t *testing.T) {
	t.Parallel()

	t.Run("returns response body as []byte", func(t *testing.T) {
		t.Parallel()

		defer gock.Off()

		endpoint := randomEndpoint()

		gock.New(endpoint).
			Get("").
			Reply(200).
			Body(strings.NewReader("hello world"))

		response, err := GET(context.Background(), endpoint).Send()

		assert.NoError(t, err)

		assert.Equal(t, []byte("hello world"), response.Bytes())
	})
}

func Test_Response_JSON(t *testing.T) {
	t.Parallel()

	t.Run("returns error if json can be parsed", func(t *testing.T) {
		t.Parallel()

		defer gock.Off()

		endpoint := randomEndpoint()

		gock.New(endpoint).
			Get("").
			Reply(200).
			Body(strings.NewReader("not json"))

		response, err := GET(context.Background(), endpoint).Send()

		assert.NoError(t, err)
		var out interface{}
		assert.NotNil(t, response.JSON(&out))
	})

	t.Run("returns response body as json", func(t *testing.T) {
		t.Parallel()

		defer gock.Off()

		endpoint := randomEndpoint()

		expected := map[string]string{"foo": "bar"}

		gock.New(endpoint).
			Get("").
			Reply(200).
			JSON(expected)

		out := make(map[string]string, 0)

		response, err := GET(context.Background(), endpoint).Send()

		assert.NoError(t, err)

		assert.NoError(t, response.JSON(&out))

		assert.Equal(t, expected, out)
	})
}
