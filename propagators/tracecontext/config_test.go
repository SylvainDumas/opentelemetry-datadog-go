package tracecontext

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_HeaderKey_setDefaultIfEmpty(t *testing.T) {
	// Check each value is empty
	var headerKey HeaderKey
	require.Empty(t, headerKey.TraceID)
	require.Empty(t, headerKey.ParentID)
	require.Empty(t, headerKey.SampledPriority)

	// Check default values applied
	headerKey.setDefaultIfEmpty()
	assert.Equal(t, DefaultTraceIDHeader, headerKey.TraceID)
	assert.Equal(t, DefaultParentIDHeader, headerKey.ParentID)
	assert.Equal(t, DefaultPriorityHeader, headerKey.SampledPriority)
}

func Test_HeaderKey_Validate(t *testing.T) {
	assert.NoError(t, (&HeaderKey{"a", "b", "c"}).Validate())
	assert.ErrorIs(t, (&HeaderKey{"a", "a", "c"}).Validate(), ErrDuplicatedHeaderKey)
	assert.ErrorIs(t, (&HeaderKey{"a", "b", "a"}).Validate(), ErrDuplicatedHeaderKey)
	assert.ErrorIs(t, (&HeaderKey{"a", "b", "b"}).Validate(), ErrDuplicatedHeaderKey)
}

func Test_Config_NewConfig(t *testing.T) {
	assert := assert.New(t)

	// Check default values applied
	if conf, err := newConfig(); assert.NoError(err) {
		test_Config_CheckDefaultValue(t, conf)
	}

	// Check invalid configuration
	_, err := newConfig(WithHeaderKey(HeaderKey{"a", "b", "a"}))
	assert.ErrorIs(err, ErrDuplicatedHeaderKey)

	var expected = "test header key"

	// Check WithHeaderKey TraceID
	if conf, err := newConfig(WithHeaderKey(HeaderKey{TraceID: expected})); assert.NoError(err) {
		assert.Equal(expected, conf.headerKey.TraceID)
		assert.Equal(DefaultParentIDHeader, conf.headerKey.ParentID)
		assert.Equal(DefaultPriorityHeader, conf.headerKey.SampledPriority)
		assert.Equal(NewHeaderConvBinary(), conf.headerValueConv)
	}

	// Check WithHeaderKey ParentID
	if conf, err := newConfig(WithHeaderKey(HeaderKey{ParentID: expected})); assert.NoError(err) {
		assert.Equal(DefaultTraceIDHeader, conf.headerKey.TraceID)
		assert.Equal(expected, conf.headerKey.ParentID)
		assert.Equal(DefaultPriorityHeader, conf.headerKey.SampledPriority)
		assert.Equal(NewHeaderConvBinary(), conf.headerValueConv)
	}

	// Check WithHeaderKey SampledPriority
	if conf, err := newConfig(WithHeaderKey(HeaderKey{SampledPriority: expected})); assert.NoError(err) {
		assert.Equal(DefaultTraceIDHeader, conf.headerKey.TraceID)
		assert.Equal(DefaultParentIDHeader, conf.headerKey.ParentID)
		assert.Equal(expected, conf.headerKey.SampledPriority)
		assert.Equal(NewHeaderConvBinary(), conf.headerValueConv)
	}

	// Check WithSampledPriorityHeader
	var expectedConv = NewHeaderConvString()
	if conf, err := newConfig(WithHeaderValueConverter(expectedConv)); assert.NoError(err) {
		assert.Equal(DefaultTraceIDHeader, conf.headerKey.TraceID)
		assert.Equal(DefaultParentIDHeader, conf.headerKey.ParentID)
		assert.Equal(DefaultPriorityHeader, conf.headerKey.SampledPriority)
		assert.Equal(expectedConv, conf.headerValueConv)
	}
}

func Test_Config_ApplyDefault(t *testing.T) {
	// Check each value is empty
	var conf config
	require.Empty(t, conf.headerKey.TraceID)
	require.Empty(t, conf.headerKey.ParentID)
	require.Empty(t, conf.headerKey.SampledPriority)
	require.Empty(t, conf.headerValueConv)

	// Check default values applied
	conf.applyDefault()
	test_Config_CheckDefaultValue(t, &conf)
}

func test_Config_CheckDefaultValue(t *testing.T, conf *config) {
	assert.Equal(t, DefaultTraceIDHeader, conf.headerKey.TraceID)
	assert.Equal(t, DefaultParentIDHeader, conf.headerKey.ParentID)
	assert.Equal(t, DefaultPriorityHeader, conf.headerKey.SampledPriority)
	assert.Equal(t, NewHeaderConvBinary(), conf.headerValueConv)
}
