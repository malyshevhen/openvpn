package client_test

import (
	"io"
	"testing"

	. "github.com/malyshevhen/openvpn/pkg/client"
	. "github.com/malyshevhen/openvpn/pkg/events"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		conn    io.ReadWriteCloser
		eventCh chan<- Event
		want    *MgmtClient
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewClient(tt.conn, tt.eventCh)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("NewClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

