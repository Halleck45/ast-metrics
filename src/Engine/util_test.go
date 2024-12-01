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

}
