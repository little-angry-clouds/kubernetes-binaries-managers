package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetLastPage(t *testing.T) {
	var flagtests = []struct {
		name     string
		input    string
		expected int
	}{
		{"value 20", "<https://api.github.com/repositories/20580498/releases?per_page=20&page=1>; rel=\"next\", <https://api.github.com/repositories/20580498/releases?per_page=20&page=20>; rel=\"last\"", 20}, // nolint: lll
		{"value 30", "<https://api.github.com/repositories/20580498/releases?per_page=20&page=1>; rel=\"next\", <https://api.github.com/repositories/20580498/releases?per_page=20&page=30>; rel=\"last\"", 30}, // nolint: lll
		{"value 2", "<https://api.github.com/repositories/20580498/releases?per_page=20&page=1>; rel=\"next\", <https://api.github.com/repositories/20580498/releases?per_page=20&page=0>; rel=\"last\"", 2},    // nolint: lll
		{"vaue 1", "<https://api.github.com/repositories/20580498/releases?per_page=20&page=1>; rel=\"next\", <https://api.github.com/repositories/20580498/releases?per_page=20&page=1>; rel=\"last\"", 1},     // nolint: lll
	}

	for _, tt := range flagtests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			actualLastPage, err := getLastPage(tt.input)
			expectedLastPage := tt.expected
			assert.Nil(t, err)
			assert.Equal(t, expectedLastPage, actualLastPage)
		})
	}
}
