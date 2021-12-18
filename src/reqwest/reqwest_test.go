package reqwest

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/IQ-tech/go-datagen"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func randomEndpoint() string {
	return fmt.Sprintf("http://localhost:500/users/%d", datagen.ID())
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
		request := GET(context.Background(), "https://api.github.com/users/poorlydefinedbehaviour/repos").
			Header("key1", "value1").
			Header("key2", "value2").
			Header("key3", "VALUE3").
			Build().
			Request()

		expected := map[string][]string{
			"key1": {"value1"},
			"key2": {"value2"},
			"key3": {"VALUE3"},
		}

		assert.EqualValues(t, expected, request.Header)
	})
}

func Test_PureRequestBuilder_Body(t *testing.T) {
	t.Parallel()

	t.Run("adds any io.Reader to request body", func(t *testing.T) {
		t.Parallel()

		payload := "hello world"

		request := POST(context.Background(), "https://api.github.com/users/poorlydefinedbehaviour/repos").
			Body(strings.NewReader(payload)).
			Build().
			Request()

		requestBody, err := ioutil.ReadAll(request.Body)

		assert.NoError(t, err)

		assert.EqualValues(t, payload, requestBody)
	})
}

func Test_ImpureRequestBuilder_Header(t *testing.T) {
	t.Parallel()

	t.Run("adds header to request", func(t *testing.T) {
		request := POST(context.Background(), "https://api.github.com/users/poorlydefinedbehaviour/repos").
			Header("key1", "value1").
			Header("key2", "value2").
			Header("key3", "VALUE3").
			Build().
			Request()

		expected := map[string][]string{
			"key1": {"value1"},
			"key2": {"value2"},
			"key3": {"VALUE3"},
		}

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

		request := POST(context.Background(), "https://api.github.com/users/poorlydefinedbehaviour/repos").
			JSON(payload).
			Build().
			Request()

		requestBody, err := ioutil.ReadAll(request.Body)

		assert.NoError(t, err)

		out := make(map[string]string)

		json.Unmarshal(requestBody, &out)

		assert.EqualValues(t, payload, out)
	})
}

func Test_ResponseBuilder_Request(t *testing.T) {
	t.Parallel()

	t.Run("returns the http request", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		endpoint := randomEndpoint()

		request := GET(ctx, endpoint).Build().Request()

		assert.Equal(t, ctx, request.Context())
		assert.Equal(t, http.MethodGet, request.Method)
		assert.Equal(t, endpoint, request.URL.String())
	})
}

func Test_ResponseBuilder_Response(t *testing.T) {
	t.Parallel()

	t.Run("returns the http response", func(t *testing.T) {
		t.Parallel()

		defer gock.Off()

		payload := "hello world"
		endpoint := randomEndpoint()

		gock.New(endpoint).
			Get("/").
			Reply(200).
			Body(strings.NewReader(payload))

		ctx := context.Background()

		response, err := GET(ctx, endpoint).Build().Response()

		assert.NoError(t, err)

		assert.Equal(t, http.StatusOK, response.StatusCode)

		responseBody, err := ioutil.ReadAll(response.Body)

		assert.NoError(t, err)

		assert.EqualValues(t, payload, responseBody)
	})
}

func Test_ResponseBuilder_Text(t *testing.T) {
	t.Parallel()

	t.Run("returns error if an error happened in the process", func(t *testing.T) {
		t.Parallel()

		builder := GET(context.Background(), "https://api.github.com/users/poorlydefinedbehaviour/repos")
		expectedErr := errors.New("some error")
		builder.err = expectedErr

		_, err := builder.Build().Text()

		assert.True(t, errors.Is(err, expectedErr))
	})

	t.Run("returns response body as text", func(t *testing.T) {
		t.Parallel()

		defer gock.Off()

		endpoint := randomEndpoint()

		gock.New(endpoint).
			Get("/").
			Reply(200).
			JSON(map[string]string{"foo": "bar"})

		body, err := GET(context.Background(), endpoint).
			Build().
			Text()

		assert.NoError(t, err)

		assert.JSONEq(t, `{"foo":"bar"}`, body)
	})
}

func Test_ResponseBuilder_Bytes(t *testing.T) {
	t.Parallel()

	t.Run("returns error if an error happened in the process", func(t *testing.T) {
		t.Parallel()

		builder := GET(context.Background(), "https://api.github.com/users/poorlydefinedbehaviour/repos")
		expectedErr := errors.New("some error")
		builder.err = expectedErr

		_, err := builder.Build().Bytes()

		assert.True(t, errors.Is(err, expectedErr))
	})

	t.Run("returns response body as []byte", func(t *testing.T) {
		t.Parallel()

		defer gock.Off()

		endpoint := randomEndpoint()

		gock.New(endpoint).
			Get("/").
			Reply(200).
			Body(strings.NewReader("hello world"))

		body, err := GET(context.Background(), endpoint).
			Build().
			Bytes()

		assert.NoError(t, err)

		assert.Equal(t, []byte("hello world"), body)
	})
}

func Test_ResponseBuilder_JSON(t *testing.T) {
	t.Parallel()

	t.Run("returns error if an error happened in the process", func(t *testing.T) {
		t.Parallel()

		builder := GET(context.Background(), "https://api.github.com/users/poorlydefinedbehaviour/repos")
		expectedErr := errors.New("some error")
		builder.err = expectedErr

		var out interface{}

		err := builder.Build().JSON(&out)

		assert.True(t, errors.Is(err, expectedErr))
	})

	t.Run("adds accept header to request", func(t *testing.T) {
		t.Parallel()

		defer gock.Off()

		endpoint := randomEndpoint()

		gock.New(endpoint).
			Get("/").
			Reply(200).
			JSON(map[string]string{})

		builder := GET(context.Background(), endpoint).Build()

		out := make(map[string]string, 0)
		err := builder.JSON(&out)

		assert.NoError(t, err)

		request := builder.Request()

		assert.Equal(t, "application/json", request.Header.Get("accept"))
	})

	t.Run("returns response body as json", func(t *testing.T) {
		t.Parallel()

		defer gock.Off()

		endpoint := randomEndpoint()

		expected := map[string]string{"foo": "bar"}

		gock.New(endpoint).
			Get("/").
			Reply(200).
			JSON(expected)

		out := make(map[string]string, 0)

		err := GET(context.Background(), endpoint).
			Build().
			JSON(&out)

		assert.NoError(t, err)

		assert.Equal(t, expected, out)
	})
}
