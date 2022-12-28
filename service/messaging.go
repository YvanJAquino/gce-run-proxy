package main

import (
	"encoding/json"
	"io"
)

type Message struct {
	Message string `json:"message,omitempty"`
}

func (m *Message) EncodeTo(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "/t")
	err := enc.Encode(m)
	if err != nil {
		return err
	}
	return nil
}

func (m *Message) WriteTo(w io.Writer) (int64, error) {
	bytes := m.MustMarshal()
	n, err := w.Write(bytes)
	return int64(n), err
}

func (m *Message) MustWriteTo(w io.Writer) int64 {
	n, err := m.WriteTo(w)
	if err != nil {
		panic(err)
	}
	return n
}

func (m *Message) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Message) MustMarshal() []byte {
	bytes, err := m.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return bytes
}

const (
	AuthorizationHeader string = "Authorization"
)

var (
	Forbidden *Message = &Message{Message: "403 FORBIDDEN - check Authorization header"}
	StatusOK  *Message = &Message{Message: "200 STATUS OK"}
)

var (
	ForbiddenBinary []byte = Forbidden.MustMarshal()
	StatusOKBinary  []byte = StatusOK.MustMarshal()
)

type PrecompiledMessage []byte

func (m PrecompiledMessage) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(m)
	return int64(n), err
}

func (m PrecompiledMessage) MustWriteTo(w io.Writer) int64 {
	n, err := m.WriteTo(w)
	if err != nil {
		panic(err)
	}
	return n
}

var (
	ForbiddenPrecompiled PrecompiledMessage = Forbidden.MustMarshal()
	StatusOKPrecompiled  PrecompiledMessage = StatusOK.MustMarshal()
)
