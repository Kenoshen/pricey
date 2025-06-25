package pricey

import (
	"context"
	"testing"
)

func TestFirebaseStoreImplementation(t *testing.T) {
	fb, _ := NewFirebase(context.Background(), OrgGroupExtractorConfig("orgId", "groupId"), nil)
	// if this line does not compile, then there is a problem with the implementation of the interface
	_ = New(fb, fb, nil)
}
