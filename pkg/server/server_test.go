package server_test

import (
	"testing"

	. "github.com/malyshevhen/openvpn/pkg/server"
)

func TestListen(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		laddr   string
		want    *MgmtListener
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := Listen(tt.laddr)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Listen() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Listen() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("Listen() = %v, want %v", got, tt.want)
			}
		})
	}
}
