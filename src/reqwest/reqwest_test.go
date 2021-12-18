package reqwest

import (
	"context"
	"net/http"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func Test_Request(t *testing.T) {
	t.Parallel()

	t.Run("returns the http request", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		endpoint := "https://api.github.com/users/poorlydefinedbehaviour/repos"

		request := GET(ctx, endpoint).Build().Request()

		assert.Equal(t, ctx, request.Context())
		assert.Equal(t, http.MethodGet, request.Method)
		assert.Equal(t, endpoint, request.URL.String())
	})
}

func Test_Header(t *testing.T) {
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

func Test_Text(t *testing.T) {
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

		const endpoint = "http://foo.com"

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

func Test_JSON(t *testing.T) {
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

		const endpoint = "http://foo.com"

		gock.New(endpoint).
			Get("/").
			Reply(200).
			JSON(map[string]string{})

		builder := GET(context.Background(), endpoint).
			Build()

		out := make(map[string]string, 0)
		err := builder.JSON(&out)

		assert.NoError(t, err)

		request := builder.Request()

		assert.Equal(t, "application/json", request.Header.Get("accept"))
	})

	t.Run("returns response body as json", func(t *testing.T) {
		t.Parallel()

		defer gock.Off()

		const endpoint = "http://foo.com"

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
