package csvlib

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	errTest1 = errors.New("test error 1")
	errTest2 = errors.New("test error 2")
	errTest3 = errors.New("test error 3")

	errCell1 = NewCellError(errTest1, 0, "column-1")
	errCell2 = NewCellError(errTest2, 1, "column-2")

	errRow1 = func() *RowErrors {
		e := NewRowErrors(1, 11)
		e.Add(errTest1)
		e.Add(errCell1)
		return e
	}()
	errRow2 = func() *RowErrors {
		e := NewRowErrors(2, 22)
		e.Add(errTest2)
		e.Add(errTest3)
		e.Add(errCell2)
		return e
	}()
)

func TestErrors(t *testing.T) {
	e := NewErrors()
	assert.Equal(t, 0, e.TotalRow())
	assert.Equal(t, "", e.Error())
	assert.False(t, e.HasError())
	assert.Equal(t, 0, e.TotalError())
	assert.Equal(t, 0, e.TotalRowError())
	assert.Equal(t, 0, e.TotalCellError())

	e.Add(errTest1)
	e.Add(errRow1)
	assert.Equal(t, 1, e.TotalRowError())
	assert.Equal(t, 1, e.TotalCellError())
	assert.Equal(t, 3, e.TotalError()) // errRow1 has 2 inner errors
}

func TestErrors_Is(t *testing.T) {
	e := NewErrors()
	assert.False(t, errors.Is(e, errTest1))
	assert.False(t, errors.Is(e, errRow1))
	assert.False(t, errors.Is(e, errCell1))

	e.Add(errTest1)
	e.Add(errRow1)
	assert.True(t, errors.Is(e, errTest1))
	assert.True(t, errors.Is(e, errRow1))
	assert.True(t, errors.Is(e, errCell1)) // as errRow1 contains errCell1
	assert.False(t, errors.Is(e, errCell2))
	assert.False(t, errors.Is(e, errRow2))
}

func TestRowErrors(t *testing.T) {
	e := NewRowErrors(1, 11)
	assert.Equal(t, 1, e.Row())
	assert.Equal(t, 11, e.Line())
	assert.Equal(t, "", e.Error())
	assert.False(t, e.HasError())
	assert.Equal(t, 0, e.TotalError())
	assert.Equal(t, 0, e.TotalCellError())

	e.Add(errTest1)
	e.Add(errCell1)
	assert.Equal(t, 2, e.TotalError())
	assert.Equal(t, "test error 1, test error 1", e.Error())
	assert.Equal(t, 1, e.TotalCellError())
}

func TestRowErrors_Is(t *testing.T) {
	e := NewRowErrors(1, 11)
	assert.False(t, errors.Is(e, errTest1))
	assert.False(t, errors.Is(e, errRow1))
	assert.False(t, errors.Is(e, errCell1))

	e.Add(errTest1)
	e.Add(errCell1)
	assert.True(t, errors.Is(e, errTest1))
	assert.True(t, errors.Is(e, errCell1))
	assert.False(t, errors.Is(e, errCell2))
}

func TestCellError(t *testing.T) {
	e1 := NewCellError(nil, 1, "column-1")
	assert.Equal(t, "column-1", e1.Header())
	assert.Equal(t, 1, e1.Column())
	assert.Equal(t, "", e1.Error())
	assert.False(t, e1.HasError())

	e2 := NewCellError(errTest2, 2, "column-2")
	assert.Equal(t, errTest2.Error(), e2.Error())
	assert.Equal(t, "", e2.LocalizationKey())

	e2.SetLocalizationKey("local-key")
	assert.Equal(t, "local-key", e2.LocalizationKey())

	_ = e2.WithParam("k", 1)
	assert.Equal(t, 1, e2.fields["k"])
}

func TestCellError_Is(t *testing.T) {
	assert.False(t, errors.Is(NewCellError(nil, 1, "column-1"), errTest1))
	assert.False(t, errors.Is(NewCellError(errTest1, 1, "column-1"), errTest2))
	assert.True(t, errors.Is(NewCellError(errTest1, 1, "column-1"), errTest1))
	assert.True(t, errors.Is(NewCellError(errRow1, 1, "column-1"), errTest1))
	assert.True(t, errors.Is(NewCellError(errRow1, 1, "column-1"), errCell1))
}
