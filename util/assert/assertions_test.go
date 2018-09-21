/*
Sniperkit-Bot
- Status: analyzed
*/

package assert_test

import (
	"testing"

	"github.com/sniperkit/snk.fork.corestoreio-pkg/util/assert"
)

func TestObjectsAreEqual(t *testing.T) {
	assert.True(t, assert.ObjectsAreEqual(1, 1))
	assert.False(t, assert.ObjectsAreEqual(0, false))
}
