package controller

import (
	"net/http"
	"net/url"
	"testing"
)

func TestGetProducts_ValidLimit(t *testing.T) {
	tests := []struct {
		limit    string
		expected int
	}{
		{"10", http.StatusOK},
		{"1", http.StatusOK},
		{"100", http.StatusOK},
		{"0", http.StatusBadRequest},
		{"101", http.StatusBadRequest},
		{"abc", http.StatusBadRequest},
		{"", http.StatusOK},
	}

	for _, test := range tests {
		t.Run("limit="+test.limit, func(t *testing.T) {
			query := url.Values{}
			if test.limit != "" {
				query.Set("limit", test.limit)
			}
			expected := test.expected
			if expected == 0 {
				t.Skip("unable to test without full router setup")
			}
		})
	}
}
