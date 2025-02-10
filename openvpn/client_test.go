package openvpn_test

import (
	"io"
	"testing"

	"github.com/malyshevhen/openvpn/openvpn"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		conn    io.ReadWriter
		eventCh chan<- openvpn.Event
		want    *openvpn.MgmtClient
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := openvpn.NewClient(tt.conn, tt.eventCh)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("NewClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

