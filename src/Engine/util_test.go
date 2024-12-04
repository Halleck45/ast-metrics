package Engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNamespacesParsing(t *testing.T) {
	t.Run("Should correctly reduce depth of namespace", func(t *testing.T) {
		assert.Equal(t, "abcd/def", ReduceDepthOfNamespace("abcd/def/ghi/lkm", 2))
		assert.Equal(t, "abcd", ReduceDepthOfNamespace("abcd/def/ghi/lkm", 1))
		assert.Equal(t, "abcd.def.ghi", ReduceDepthOfNamespace("abcd.def.ghi.lkm", 3))
		assert.Equal(t, "abcd.def", ReduceDepthOfNamespace("abcd.def", 3))
		assert.Equal(t, "abcd", ReduceDepthOfNamespace("abcd/def.ghi.lkm", 1))
		assert.Equal(t, "abcd", ReduceDepthOfNamespace("abcd", 2))
	})

	t.Run("Should avoid github.com namespace", func(t *testing.T) {
		assert.Equal(t, "github.com/test/test", ReduceDepthOfNamespace("github.com/test/test/test/test", 2))
		assert.Equal(t, "github.com/test", ReduceDepthOfNamespace("github.com/test/test/test/test", 1))
		assert.Equal(t, "github.com/test/test/test", ReduceDepthOfNamespace("github.com/test.test.test.test", 3))
		assert.Equal(t, "github.com/test.test", ReduceDepthOfNamespace("github.com/test.test", 3))
		assert.Equal(t, "github.com/test", ReduceDepthOfNamespace("github.com/test/test.test.test", 1))
		assert.Equal(t, "github.com/test", ReduceDepthOfNamespace("github.com/test", 2))
	})

}
