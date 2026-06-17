package webrtc

import "errors"

var (
	ErrRoomNotFound = errors.New("room not found")
	ErrRoomFull     = errors.New("room is full")
	ErrInvalidSDP   = errors.New("invalid SDP")
	ErrInvalidCandidate = errors.New("invalid ICE candidate")
)
