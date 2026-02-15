package idem

import "errors"

var (
	ErrInvalidScope       = errors.New("invalid scope")
	ErrInvalidKey         = errors.New("invalid key")
	ErrInvalidFingerprint = errors.New("invalid fingerprint")
)
