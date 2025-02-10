package events

import (
	"bytes"
	"fmt"
	"strconv"
)

const (
	eventSep = ":"
	fieldSep = ","

	byteCountEventKW        = "BYTECOUNT"
	byteCountCliEventKW     = "BYTECOUNT_CLI"
	echoEventKW             = "ECHO"
	fatalEventKW            = "FATAL"
	holdEventKW             = "HOLD"
	stateEventKW            = "STATE"
	passwordEventKW         = "PASSWORD"
	clientConnectEventKW    = "CLIENT:CONNECT"    // TODO: not implemented
	clientDisconnectEventKW = "CLIENT:DISCONNECT" // TODO: not implemented
	clientReauthEventKW     = "CLIENT:REAUTH"     // TODO: not implemented
	clientGeneralEventKW    = "CLIENT"            // TODO: not implemented
	infoEventKW             = "INFO"              // TODO: not implemented
	logEventKW              = "LOG"               // TODO: not implemented
	needOkEventKW           = "NEED-OK"           // TODO: not implemented
	needStrEventKW          = "NEED-STR"          // TODO: not implemented
	pkcs11IdCountEventKW    = "PKCS11ID-COUNT"    // TODO: not implemented
	pkcs11IdEntryEventKW    = "PKCS11ID-ENTRY"    // TODO: not implemented
)

type Event interface {
	String() string
}

// UnknownEvent represents an event of a type that this package doesn't
// know about.
//
// Future versions of this library may learn about new event types, so a
// caller should exercise caution when making use of events of this type
// to access unsupported behavior. Backward-compatibility is *not*
// guaranteed for events of this type.
type UnknownEvent interface {
	Type() string
	Body() string
	String() string
}

func newUnknownEvent(keyword []byte, body []byte) UnknownEvent {
	return &unknownEvent{
		keyword,
		body,
	}
}

type unknownEvent struct {
	keyword []byte
	body    []byte
}

func (e *unknownEvent) Type() string {
	return string(e.keyword)
}

func (e *unknownEvent) Body() string {
	return string(e.body)
}

func (e *unknownEvent) String() string {
	return fmt.Sprintf("%s: %s", e.keyword, e.body)
}

// MalformedEvent represents a message from the OpenVPN process that is
// presented as an event but does not comply with the expected event syntax.
//
// Events of this type should never be seen but robust callers will accept
// and ignore them, possibly generating some kind of debugging message.
//
// One reason for potentially seeing events of this type is when the target
// program is actually not an OpenVPN process at all, but in fact this client
// has been connected to a different sort of server by mistake.
type MalformedEvent interface {
	String() string
}

func newMalformedEvent(raw []byte) MalformedEvent {
	return &malformedEvent{raw}
}

type malformedEvent struct {
	raw []byte
}

func (e *malformedEvent) String() string {
	return fmt.Sprintf("Malformed Event %q", e.raw)
}

// HoldEvent is a notification that the OpenVPN process is in a management
// hold and will not continue connecting until the hold is released, e.g.
// by calling client.HoldRelease()
type HoldEvent interface {
	String() string
}

func newHoldEvent(body []byte) HoldEvent {
	return &holdEvent{body}
}

type holdEvent struct {
	body []byte
}

func (e *holdEvent) String() string {
	return string(e.body)
}

// StateEvent is a notification of a change of connection state. It can be
// used, for example, to detect if the OpenVPN connection has been interrupted
// and the OpenVPN process is attempting to reconnect.
type StateEvent interface {
	RawTimestamp() string
	NewState() string
	Description() string

	// LocalTunnelAddr returns the IP address of the local interface within
	// the tunnel, as a string that can be parsed using net.ParseIP.
	//
	// This field is only populated for events whose NewState returns
	// either ASSIGN_IP or CONNECTED.
	LocalTunnelAddr() string

	// RemoteAddr returns the non-tunnel IP address of the remote
	// system that has connected to the local OpenVPN process.
	//
	// This field is only populated for events whose NewState returns
	// CONNECTED.
	RemoteAddr() string
	String() string
}

// TODO: Docs
func NewStateEvent(body []byte) StateEvent {
	return &stateEvent{
		body: body,
	}
}

type stateEvent struct {
	body []byte

	// bodyParts is populated only on first request, giving us the
	// separate comma-separated elements of the message. Not all
	// fields are populated for all states.
	bodyParts [][]byte
}

func (e *stateEvent) RawTimestamp() string {
	parts := e.parts()
	return string(parts[0])
}

func (e *stateEvent) NewState() string {
	parts := e.parts()
	return string(parts[1])
}

func (e *stateEvent) Description() string {
	parts := e.parts()
	return string(parts[2])
}

func (e *stateEvent) LocalTunnelAddr() string {
	parts := e.parts()
	if len(parts) > 8 {
		return string(parts[8])
	}
	return string(parts[3])
}

func (e *stateEvent) RemoteAddr() string {
	parts := e.parts()
	return string(parts[4])
}

func (e *stateEvent) String() string {
	newState := e.NewState()
	switch newState {
	case "ASSIGN_IP":
		return fmt.Sprintf("%s: %s", newState, e.LocalTunnelAddr())
	case "CONNECTED":
		return fmt.Sprintf("%s: %s", newState, e.RemoteAddr())
	default:
		desc := e.Description()
		if desc != "" {
			return fmt.Sprintf("%s: %s", newState, desc)
		} else {
			return newState
		}
	}
}

func (e *stateEvent) parts() [][]byte {
	if e.bodyParts == nil {
		e.bodyParts = bytes.SplitN(e.body, []byte(fieldSep), 9)

		// Prevent crash if the server has sent us a malformed
		// status message. This should never actually happen if
		// the server is behaving itself.
		if len(e.bodyParts) < 8 {
			expanded := make([][]byte, 8)
			copy(expanded, e.bodyParts)
			e.bodyParts = expanded
		}
	}
	return e.bodyParts
}

// EchoEvent is emitted by an OpenVPN process running in client mode when
// an "echo" command is pushed to it by the server it has connected to.
//
// The format of the echo message is free-form, since this message type is
// intended to pass application-specific data from the server-side config
// into whatever client is consuming the management protocol.
//
// This event is emitted only if the management client has turned on events
// of this type using client.SetEchoEvents(true)
type EchoEvent interface {
	RawTimestamp() string
	Message() string
	String() string
}

func newEchoEvent(body []byte) EchoEvent {
	return &echoEvent{body}
}

type echoEvent struct {
	body []byte
}

func (e *echoEvent) RawTimestamp() string {
	sepIndex := bytes.Index(e.body, []byte(fieldSep))
	if sepIndex == -1 {
		return ""
	}
	return string(e.body[:sepIndex])
}

func (e *echoEvent) Message() string {
	sepIndex := bytes.Index(e.body, []byte(fieldSep))
	if sepIndex == -1 {
		return ""
	}
	return string(e.body[sepIndex+1:])
}

func (e *echoEvent) String() string {
	return fmt.Sprintf("ECHO: %s", e.Message())
}

// ByteCountEvent represents a periodic snapshot of data transfer in bytes
// on a VPN connection.
//
// For OpenVPN *servers*, events are emitted for each client and the method
// ClientId identifies that client.
//
// For other OpenVPN modes, events are emitted only once per interval for the
// single connection managed by the target process, and ClientId returns
// the empty string.
type ByteCountEvent interface {
	ClientId() string
	BytesIn() int
	BytesOut() int
	String() string
}

func newByteCountEvent(hasClient bool, body []byte) ByteCountEvent {
	return &byteCountEvent{
		hasClient: hasClient,
		body:      body,
	}
}

type byteCountEvent struct {
	hasClient bool
	body      []byte

	// populated on first call to parts()
	bodyParts [][]byte
}

func (e *byteCountEvent) ClientId() string {
	if !e.hasClient {
		return ""
	}

	return string(e.parts()[0])
}

func (e *byteCountEvent) BytesIn() int {
	index := 0
	if e.hasClient {
		index = 1
	}
	str := string(e.parts()[index])
	val, _ := strconv.Atoi(str)
	// Ignore error, since this should never happen if OpenVPN is
	// behaving itself.
	return val
}

func (e *byteCountEvent) BytesOut() int {
	index := 1
	if e.hasClient {
		index = 2
	}
	str := string(e.parts()[index])
	val, _ := strconv.Atoi(str)
	// Ignore error, since this should never happen if OpenVPN is
	// behaving itself.
	return val
}

func (e *byteCountEvent) String() string {
	if e.hasClient {
		return fmt.Sprintf("Client %s: %d in, %d out", e.ClientId(), e.BytesIn(), e.BytesOut())
	} else {
		return fmt.Sprintf("%d in, %d out", e.BytesIn(), e.BytesOut())
	}
}

func (e *byteCountEvent) parts() [][]byte {
	if e.bodyParts == nil {
		e.bodyParts = bytes.SplitN(e.body, []byte(fieldSep), 4)

		wantCount := 2
		if e.hasClient {
			wantCount = 3
		}

		// Prevent crash if the server has sent us a malformed
		// message. This should never actually happen if the
		// server is behaving itself.
		if len(e.bodyParts) < wantCount {
			expanded := make([][]byte, wantCount)
			copy(expanded, e.bodyParts)
			e.bodyParts = expanded
		}
	}
	return e.bodyParts
}

// PasswordEvent represents a message from the OpenVPN process asking for
// authentication data, such as username and password.
type PasswordEvent interface {
	String() string
}

func newPasswordEvent(body []byte) PasswordEvent {
	return &passwordEvent{body}
}

type passwordEvent struct {
	body []byte
}

func (e *passwordEvent) String() string {
	return fmt.Sprintf("PASSWORD: %s", string(e.body))
}

// FatalEvent represents a message from the OpenVPN process before exiting.
type FatalEvent interface {
	String() string
}

func newFatalEvent(body []byte) FatalEvent {
	return &fatalEvent{body}
}

type fatalEvent struct {
	body []byte
}

func (e *fatalEvent) String() string {
	return fmt.Sprintf("FATAL: %s", string(e.body))
}

func UpgradeEvent(raw []byte) Event {
	splitIdx := bytes.Index(raw, []byte(eventSep))
	if splitIdx == -1 {
		// Should never happen, but we'll handle it robustly if it does.
		return newMalformedEvent(raw)
	}

	keyword := raw[:splitIdx]
	body := raw[splitIdx+1:]

	switch string(keyword) {
	case stateEventKW:
		return NewStateEvent(body)
	case holdEventKW:
		return newHoldEvent(body)
	case echoEventKW:
		return newEchoEvent(body)
	case byteCountEventKW:
		return newByteCountEvent(false, body)
	case byteCountCliEventKW:
		return newByteCountEvent(true, body)
	case passwordEventKW:
		return newPasswordEvent(body)
	case fatalEventKW:
		return newFatalEvent(body)
	default:
		return newUnknownEvent(keyword, body)
	}
}
