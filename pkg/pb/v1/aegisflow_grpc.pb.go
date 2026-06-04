package pbv1

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AgentOSGatewayClient interface {
	CheckAction(ctx context.Context, in *CheckActionRequest, opts ...grpc.CallOption) (*CheckActionResponse, error)
	IssueCredential(ctx context.Context, in *IssueCredentialRequest, opts ...grpc.CallOption) (*IssueCredentialResponse, error)
	RecordEvidence(ctx context.Context, in *RecordEvidenceRequest, opts ...grpc.CallOption) (*RecordEvidenceResponse, error)
}

type agentOSGatewayClient struct {
	cc grpc.ClientConnInterface
}

func NewAgentOSGatewayClient(cc grpc.ClientConnInterface) AgentOSGatewayClient {
	return &agentOSGatewayClient{cc}
}

func (c *agentOSGatewayClient) CheckAction(ctx context.Context, in *CheckActionRequest, opts ...grpc.CallOption) (*CheckActionResponse, error) {
	out := new(CheckActionResponse)
	err := c.cc.Invoke(ctx, "/aegisflow.v1.AgentOSGateway/CheckAction", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentOSGatewayClient) IssueCredential(ctx context.Context, in *IssueCredentialRequest, opts ...grpc.CallOption) (*IssueCredentialResponse, error) {
	out := new(IssueCredentialResponse)
	err := c.cc.Invoke(ctx, "/aegisflow.v1.AgentOSGateway/IssueCredential", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *agentOSGatewayClient) RecordEvidence(ctx context.Context, in *RecordEvidenceRequest, opts ...grpc.CallOption) (*RecordEvidenceResponse, error) {
	out := new(RecordEvidenceResponse)
	err := c.cc.Invoke(ctx, "/aegisflow.v1.AgentOSGateway/RecordEvidence", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

type AgentOSGatewayServer interface {
	CheckAction(context.Context, *CheckActionRequest) (*CheckActionResponse, error)
	IssueCredential(context.Context, *IssueCredentialRequest) (*IssueCredentialResponse, error)
	RecordEvidence(context.Context, *RecordEvidenceRequest) (*RecordEvidenceResponse, error)
	mustEmbedUnimplementedAgentOSGatewayServer()
}

type UnimplementedAgentOSGatewayServer struct{}

func (UnimplementedAgentOSGatewayServer) CheckAction(context.Context, *CheckActionRequest) (*CheckActionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckAction not implemented")
}

func (UnimplementedAgentOSGatewayServer) IssueCredential(context.Context, *IssueCredentialRequest) (*IssueCredentialResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method IssueCredential not implemented")
}

func (UnimplementedAgentOSGatewayServer) RecordEvidence(context.Context, *RecordEvidenceRequest) (*RecordEvidenceResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RecordEvidence not implemented")
}

func (UnimplementedAgentOSGatewayServer) mustEmbedUnimplementedAgentOSGatewayServer() {}

func RegisterAgentOSGatewayServer(s grpc.ServiceRegistrar, srv AgentOSGatewayServer) {
	s.RegisterService(&AgentOSGateway_ServiceDesc, srv)
}

func _AgentOSGateway_CheckAction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckActionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentOSGatewayServer).CheckAction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/aegisflow.v1.AgentOSGateway/CheckAction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentOSGatewayServer).CheckAction(ctx, req.(*CheckActionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentOSGateway_IssueCredential_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IssueCredentialRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentOSGatewayServer).IssueCredential(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/aegisflow.v1.AgentOSGateway/IssueCredential",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentOSGatewayServer).IssueCredential(ctx, req.(*IssueCredentialRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _AgentOSGateway_RecordEvidence_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RecordEvidenceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AgentOSGatewayServer).RecordEvidence(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/aegisflow.v1.AgentOSGateway/RecordEvidence",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(AgentOSGatewayServer).RecordEvidence(ctx, req.(*RecordEvidenceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var AgentOSGateway_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "aegisflow.v1.AgentOSGateway",
	HandlerType: (*AgentOSGatewayServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CheckAction",
			Handler:    _AgentOSGateway_CheckAction_Handler,
		},
		{
			MethodName: "IssueCredential",
			Handler:    _AgentOSGateway_IssueCredential_Handler,
		},
		{
			MethodName: "RecordEvidence",
			Handler:    _AgentOSGateway_RecordEvidence_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/aegisflow.proto",
}
