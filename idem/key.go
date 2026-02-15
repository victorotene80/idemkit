package idem

import (
	"strings"
)

type Scope string

func (s Scope) String() string { return string(s) }

func (s Scope) Normalize() Scope {
	return Scope(strings.ToLower(strings.TrimSpace(string(s))))
}

func (s Scope) Valid() bool {
	v := strings.TrimSpace(string(s))
	if v == "" || len(v) > 64 {
		return false
	}
	return validIdent(v)
}

type Key struct {
	scope Scope
	id    string
}

func NewKey(scope Scope, id string) (Key, error) {
	scope = scope.Normalize()
	id = normalizeID(id)

	if !scope.Valid() {
		return Key{}, ErrInvalidScope
	}
	if !validID(id) {
		return Key{}, ErrInvalidKey
	}
	return Key{scope: scope, id: id}, nil
}

func MustKey(scope Scope, id string) Key {
	k, err := NewKey(scope, id)
	if err != nil {
		panic(err)
	}
	return k
}

func (k Key) Scope() Scope { return k.scope }
func (k Key) ID() string   { return k.id }

func (k Key) String() string { return k.scope.String() + ":" + k.id }

func (k Key) Normalize() (Key, error) {
	return NewKey(k.scope, k.id)
}

func (k Key) Valid() bool {
	_, err := NewKey(k.scope, k.id)
	return err == nil
}

func normalizeID(s string) string {
	return strings.TrimSpace(s)
}

func validID(s string) bool {
	if s == "" || len(s) > 128 {
		return false
	}
	return validIdent(s)
}

func validIdent(s string) bool {
	for _, r := range s {
		if (r >= 'A' && r <= 'Z') ||
			(r >= 'a' && r <= 'z') ||
			(r >= '0' && r <= '9') ||
			r == '_' || r == '-' || r == '.' || r == ':' {
			continue
		}
		return false
	}
	return true
}
