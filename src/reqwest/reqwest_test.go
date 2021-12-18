package reqwest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
