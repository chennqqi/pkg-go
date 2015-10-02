// Code generated by protoc-gen-go.
// source: http/protolog.proto
// DO NOT EDIT!

package pkghttp

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "go.pedge.io/google-protobuf"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type Call struct {
	Method         string                    `protobuf:"bytes,1,opt,name=method" json:"method,omitempty"`
	Path           string                    `protobuf:"bytes,2,opt,name=path" json:"path,omitempty"`
	Query          map[string]string         `protobuf:"bytes,3,rep,name=query" json:"query,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	RequestHeader  map[string]string         `protobuf:"bytes,4,rep,name=request_header" json:"request_header,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	RequestForm    map[string]string         `protobuf:"bytes,5,rep,name=request_form" json:"request_form,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	ResponseHeader map[string]string         `protobuf:"bytes,6,rep,name=response_header" json:"response_header,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	StatusCode     uint32                    `protobuf:"varint,7,opt,name=status_code" json:"status_code,omitempty"`
	Duration       *google_protobuf.Duration `protobuf:"bytes,8,opt,name=duration" json:"duration,omitempty"`
	Error          string                    `protobuf:"bytes,9,opt,name=error" json:"error,omitempty"`
}

func (m *Call) Reset()         { *m = Call{} }
func (m *Call) String() string { return proto.CompactTextString(m) }
func (*Call) ProtoMessage()    {}

func (m *Call) GetQuery() map[string]string {
	if m != nil {
		return m.Query
	}
	return nil
}

func (m *Call) GetRequestHeader() map[string]string {
	if m != nil {
		return m.RequestHeader
	}
	return nil
}

func (m *Call) GetRequestForm() map[string]string {
	if m != nil {
		return m.RequestForm
	}
	return nil
}

func (m *Call) GetResponseHeader() map[string]string {
	if m != nil {
		return m.ResponseHeader
	}
	return nil
}

func (m *Call) GetDuration() *google_protobuf.Duration {
	if m != nil {
		return m.Duration
	}
	return nil
}

type ServerCouldNotStart struct {
	Error string `protobuf:"bytes,1,opt,name=error" json:"error,omitempty"`
}

func (m *ServerCouldNotStart) Reset()         { *m = ServerCouldNotStart{} }
func (m *ServerCouldNotStart) String() string { return proto.CompactTextString(m) }
func (*ServerCouldNotStart) ProtoMessage()    {}

type ServerStarting struct {
}

func (m *ServerStarting) Reset()         { *m = ServerStarting{} }
func (m *ServerStarting) String() string { return proto.CompactTextString(m) }
func (*ServerStarting) ProtoMessage()    {}

type ServerFinished struct {
	Error    string                    `protobuf:"bytes,1,opt,name=error" json:"error,omitempty"`
	Duration *google_protobuf.Duration `protobuf:"bytes,2,opt,name=duration" json:"duration,omitempty"`
}

func (m *ServerFinished) Reset()         { *m = ServerFinished{} }
func (m *ServerFinished) String() string { return proto.CompactTextString(m) }
func (*ServerFinished) ProtoMessage()    {}

func (m *ServerFinished) GetDuration() *google_protobuf.Duration {
	if m != nil {
		return m.Duration
	}
	return nil
}
