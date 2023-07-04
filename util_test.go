package csvlib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_validateHeader(t *testing.T) {
	assert.Nil(t, validateHeader([]string{"col1", "col2", "col3"}))
	assert.Nil(t, validateHeader([]string{"col1", "col2", "Col1"}))

	assert.ErrorIs(t, validateHeader([]string{"col1", "col2", "col3 "}), ErrHeaderColumnInvalid)
	assert.ErrorIs(t, validateHeader([]string{"col1", "col2", "col1"}), ErrHeaderColumnDuplicated)
}
