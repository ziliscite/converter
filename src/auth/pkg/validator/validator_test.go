package validator

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateValidator_Success(t *testing.T) {
	validator := New()
	assert.NotNil(t, validator)
	assert.NotNil(t, validator.errs)
	require.Equal(t, make(map[string]string), validator.errs)
}

func TestCreateError_Success(t *testing.T) {
	errs := map[string]string{
		"title":   "title cannot be nil",
		"release": "invalid release date",
	}

	v := New()

	v.AddError("title", "title cannot be nil")
	v.AddError("release", "invalid release date")

	require.NotNil(t, v.errs)

	ok := v.Valid()
	assert.False(t, ok)

	assert.Equal(t, v.errs["title"], errs["title"])
	assert.Equal(t, v.errs["release"], errs["release"])
	assert.Equal(t, v.Errors(), errs)
}
