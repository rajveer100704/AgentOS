package grpcgw

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/saivedant169/AegisFlow/internal/credential"
	"github.com/saivedant169/AegisFlow/internal/envelope"
	"github.com/saivedant169/AegisFlow/internal/evidence"
	"github.com/saivedant169/AegisFlow/internal/policy"
	pbv1 "github.com/saivedant169/AegisFlow/pkg/pb/v1"
)

// Server implements pbv1.AgentOSGatewayServer.
type Server struct {
	pbv1.UnimplementedAgentOSGatewayServer
	policyEngine  *policy.Engine
	credRegistry  *credential.Registry
	evidenceChain *evidence.SessionChain
}

// NewServer creates a new gRPC server wrapper.
func NewServer(pe *policy.Engine, cr *credential.Registry, ec *evidence.SessionChain) *Server {
	return &Server{
		policyEngine:  pe,
		credRegistry:  cr,
		evidenceChain: ec,
	}
}

// CheckAction evaluates an action envelope against input policies.
func (s *Server) CheckAction(ctx context.Context, req *pbv1.CheckActionRequest) (*pbv1.CheckActionResponse, error) {
	if s.policyEngine == nil {
		return nil, errors.New("policy engine not initialized")
	}

	// Run input check
	v, err := s.policyEngine.CheckInput(req.InputText)
	if err != nil {
		return nil, fmt.Errorf("policy check input: %w", err)
	}

	resp := &pbv1.CheckActionResponse{
		Decision: "allow",
	}

	if v != nil {
		resp.Decision = string(v.Action)
		resp.PolicyName = v.PolicyName
		resp.Message = v.Message
	}

	return resp, nil
}

// IssueCredential issues a temporary credential if the tenant has permission.
func (s *Server) IssueCredential(ctx context.Context, req *pbv1.IssueCredentialRequest) (*pbv1.IssueCredentialResponse, error) {
	if s.credRegistry == nil {
		return nil, errors.New("credential registry not initialized")
	}

	credReq := credential.CredentialRequest{
		TenantID:   req.TenantId,
		Target:     req.Target,
		Capability: req.Capability,
		TTL:        30 * time.Minute,
	}

	cred, _, err := s.credRegistry.Issue(ctx, req.ProviderName, credReq, "")
	if err != nil {
		return nil, fmt.Errorf("issue credential: %w", err)
	}

	return &pbv1.IssueCredentialResponse{
		CredentialId: cred.ID,
		Status:       "active",
		ExpiresAt:    cred.ExpiresAt.Unix(),
	}, nil
}

// RecordEvidence logs the action in the evidence chain.
func (s *Server) RecordEvidence(ctx context.Context, req *pbv1.RecordEvidenceRequest) (*pbv1.RecordEvidenceResponse, error) {
	if s.evidenceChain == nil {
		return nil, errors.New("evidence chain not initialized")
	}

	env := &envelope.ActionEnvelope{
		ID:        req.SessionId + "-" + fmt.Sprintf("%d", time.Now().UnixNano()),
		Timestamp: time.Now(),
		Actor: envelope.ActorInfo{
			SessionID: req.SessionId,
		},
		Tool:   req.Action,
		Target: req.Payload,
	}

	rec, err := s.evidenceChain.Record(env)
	if err != nil {
		return nil, fmt.Errorf("record evidence: %w", err)
	}

	return &pbv1.RecordEvidenceResponse{
		RecordId: rec.Hash,
		Verified: true,
	}, nil
}

// StartServer starts the gRPC server listening on the specified port.
func StartServer(addr string, pe *policy.Engine, cr *credential.Registry, ec *evidence.SessionChain) (*grpc.Server, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	s := grpc.NewServer()
	srv := NewServer(pe, cr, ec)
	pbv1.RegisterAgentOSGatewayServer(s, srv)

	go func() {
		log.Printf("AgentOS gRPC Gateway listening on %s", addr)
		if err := s.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	return s, nil
}
