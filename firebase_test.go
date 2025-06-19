package pricey

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFirebaseStoreImplementation(t *testing.T) {
	fb, err := NewFirebase(context.Background(), OrgGroupExtractorConfig("orgId", "groupId"), nil)
	require.NoError(t, err)
	// if this line does not compile, then there is a problem with the implementation of the interface
	_ = New(fb, fb, nil)
}
