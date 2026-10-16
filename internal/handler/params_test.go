package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fc1g/kachenyatko-store-api/internal/apperr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePagination(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		wantLimit  int32
		wantOffset int32
		wantErrMsg string
	}{
		{name: "defaults", query: "", wantLimit: 20, wantOffset: 0},
		{name: "custom values", query: "limit=5&offset=10", wantLimit: 5, wantOffset: 10},
		{name: "only limit", query: "limit=50", wantLimit: 50, wantOffset: 0},
		{name: "only offset", query: "offset=40", wantLimit: 20, wantOffset: 40},
		{name: "min limit", query: "limit=1", wantLimit: 1, wantOffset: 0},
		{name: "max limit", query: "limit=100", wantLimit: 100, wantOffset: 0},
		{name: "limit not a number", query: "limit=abc", wantErrMsg: "invalid limit query param"},
		{name: "limit overflows int32", query: "limit=3000000000", wantErrMsg: "invalid limit query param"},
		{name: "offset not a number", query: "offset=abc", wantErrMsg: "invalid offset query param"},
		{name: "zero limit", query: "limit=0", wantErrMsg: "limit must be greater than 0"},
		{name: "negative limit", query: "limit=-5", wantErrMsg: "limit must be greater than 0"},
		{name: "limit above max", query: "limit=101", wantErrMsg: "limit cannot exceed 100"},
		{name: "negative offset", query: "offset=-1", wantErrMsg: "offset must not be negative"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/?"+tt.query, nil)

			limit, offset, err := parsePagination(r)

			if tt.wantErrMsg != "" {
				var appErr apperr.Error
				require.ErrorAs(t, err, &appErr)

				assert.Equal(t, tt.wantErrMsg, appErr.Message)
				return
			}

			require.NoError(t, err)

			assert.Equal(t, tt.wantLimit, limit)
			assert.Equal(t, tt.wantOffset, offset)
		})
	}
}
