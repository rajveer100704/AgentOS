package pbv1

import (
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type CheckActionRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TenantId   string `protobuf:"bytes,1,opt,name=tenant_id,json=tenantId,proto3" json:"tenant_id,omitempty"`
	Tool       string `protobuf:"bytes,2,opt,name=tool,proto3" json:"tool,omitempty"`
	Target     string `protobuf:"bytes,3,opt,name=target,proto3" json:"target,omitempty"`
	Capability string `protobuf:"bytes,4,opt,name=capability,proto3" json:"capability,omitempty"`
	InputText  string `protobuf:"bytes,5,opt,name=input_text,json=inputText,proto3" json:"input_text,omitempty"`
}

func (x *CheckActionRequest) Reset() {
	*x = CheckActionRequest{}
}

func (x *CheckActionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckActionRequest) ProtoMessage() {}

func (x *CheckActionRequest) ProtoReflect() protoreflect.Message {
	return protoimpl.X.MessageOf(x)
}

type CheckActionResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Decision   string `protobuf:"bytes,1,opt,name=decision,proto3" json:"decision,omitempty"`
	PolicyName string `protobuf:"bytes,2,opt,name=policy_name,json=policyName,proto3" json:"policy_name,omitempty"`
	Message    string `protobuf:"bytes,3,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *CheckActionResponse) Reset() {
	*x = CheckActionResponse{}
}

func (x *CheckActionResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CheckActionResponse) ProtoMessage() {}

func (x *CheckActionResponse) ProtoReflect() protoreflect.Message {
	return protoimpl.X.MessageOf(x)
}

type IssueCredentialRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TenantId     string `protobuf:"bytes,1,opt,name=tenant_id,json=tenantId,proto3" json:"tenant_id,omitempty"`
	ProviderName string `protobuf:"bytes,2,opt,name=provider_name,json=providerName,proto3" json:"provider_name,omitempty"`
	Capability   string `protobuf:"bytes,3,opt,name=capability,proto3" json:"capability,omitempty"`
	Target       string `protobuf:"bytes,4,opt,name=target,proto3" json:"target,omitempty"`
}

func (x *IssueCredentialRequest) Reset() {
	*x = IssueCredentialRequest{}
}

func (x *IssueCredentialRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IssueCredentialRequest) ProtoMessage() {}

func (x *IssueCredentialRequest) ProtoReflect() protoreflect.Message {
	return protoimpl.X.MessageOf(x)
}

type IssueCredentialResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	CredentialId string `protobuf:"bytes,1,opt,name=credential_id,json=credentialId,proto3" json:"credential_id,omitempty"`
	Status       string `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
	ExpiresAt    int64  `protobuf:"varint,3,opt,name=expires_at,json=expiresAt,proto3" json:"expires_at,omitempty"`
}

func (x *IssueCredentialResponse) Reset() {
	*x = IssueCredentialResponse{}
}

func (x *IssueCredentialResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IssueCredentialResponse) ProtoMessage() {}

func (x *IssueCredentialResponse) ProtoReflect() protoreflect.Message {
	return protoimpl.X.MessageOf(x)
}

type RecordEvidenceRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SessionId string `protobuf:"bytes,1,opt,name=session_id,json=sessionId,proto3" json:"session_id,omitempty"`
	Action    string `protobuf:"bytes,2,opt,name=action,proto3" json:"action,omitempty"`
	Payload   string `protobuf:"bytes,3,opt,name=payload,proto3" json:"payload,omitempty"`
	Hash      string `protobuf:"bytes,4,opt,name=hash,proto3" json:"hash,omitempty"`
}

func (x *RecordEvidenceRequest) Reset() {
	*x = RecordEvidenceRequest{}
}

func (x *RecordEvidenceRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RecordEvidenceRequest) ProtoMessage() {}

func (x *RecordEvidenceRequest) ProtoReflect() protoreflect.Message {
	return protoimpl.X.MessageOf(x)
}

type RecordEvidenceResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RecordId string `protobuf:"bytes,1,opt,name=record_id,json=recordId,proto3" json:"record_id,omitempty"`
	Verified bool   `protobuf:"varint,2,opt,name=verified,proto3" json:"verified,omitempty"`
}

func (x *RecordEvidenceResponse) Reset() {
	*x = RecordEvidenceResponse{}
}

func (x *RecordEvidenceResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RecordEvidenceResponse) ProtoMessage() {}

func (x *RecordEvidenceResponse) ProtoReflect() protoreflect.Message {
	return protoimpl.X.MessageOf(x)
}
