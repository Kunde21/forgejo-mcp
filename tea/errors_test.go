package tea

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestValidateConnectionErrors(t *testing.T) {
	wrapper := &GiteaWrapper{}
	err := wrapper.Ping(context.Background())
	if err == nil {
		t.Fatal("Ping() error = nil, want error")
	}

	wantErr := "wrapper not initialized"
	if !cmp.Equal(wantErr, err.Error()) {
		t.Error(cmp.Diff(wantErr, err.Error()))
	}
}
