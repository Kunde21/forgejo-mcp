package server

import (
	"testing"

	remote "github.com/Kunde21/forgejo-mcp/remote/gitea"
	"github.com/sirupsen/logrus"
)

// TestHandlerDependencyInjection tests that handlers properly inject dependencies
func TestHandlerDependencyInjection(t *testing.T) {
	logger := logrus.New()
	mockClient := &remote.MockGiteaClient{}

	// Test PR handler dependency injection
	prHandler := NewSDKPRListHandler(logger, mockClient)
	if prHandler == nil {
		t.Fatal("NewSDKPRListHandler should return a valid handler")
	}
	if prHandler.logger == nil {
		t.Error("PR handler should have logger injected")
	}
	if prHandler.client == nil {
		t.Error("PR handler should have client injected")
	}

	// Test repository handler dependency injection
	repoHandler := NewSDKRepositoryHandler(logger, mockClient)
	if repoHandler == nil {
		t.Fatal("NewSDKRepositoryHandler should return a valid handler")
	}
	if repoHandler.logger == nil {
		t.Error("Repository handler should have logger injected")
	}
	if repoHandler.client == nil {
		t.Error("Repository handler should have client injected")
	}

	// Test issue handler dependency injection
	issueHandler := NewSDKIssueListHandler(logger, mockClient)
	if issueHandler == nil {
		t.Fatal("NewSDKIssueListHandler should return a valid handler")
	}
	if issueHandler.logger == nil {
		t.Error("Issue handler should have logger injected")
	}
	if issueHandler.client == nil {
		t.Error("Issue handler should have client injected")
	}
}

// TestHandlerDependencyIsolation tests that handlers are properly isolated with their dependencies
func TestHandlerDependencyIsolation(t *testing.T) {
	logger1 := logrus.New()
	logger2 := logrus.New()
	mockClient1 := &remote.MockGiteaClient{}
	mockClient2 := &remote.MockGiteaClient{}

	// Create handlers with different dependencies
	handler1 := NewSDKPRListHandler(logger1, mockClient1)
	handler2 := NewSDKPRListHandler(logger2, mockClient2)

	// Verify they have different dependencies
	if handler1.logger == handler2.logger {
		t.Error("Handlers should have isolated logger dependencies")
	}
	if handler1.client == handler2.client {
		t.Error("Handlers should have isolated client dependencies")
	}
}

// TestHandlerInitialization tests proper handler initialization
func TestHandlerInitialization(t *testing.T) {
	logger := logrus.New()
	mockClient := &remote.MockGiteaClient{}

	tests := []struct {
		name    string
		logger  *logrus.Logger
		client  remote.GiteaClientInterface
		wantNil bool
	}{
		{
			name:    "valid initialization",
			logger:  logger,
			client:  mockClient,
			wantNil: false,
		},
		{
			name:    "nil logger",
			logger:  nil,
			client:  mockClient,
			wantNil: false, // Should still create handler, but logger will be nil
		},
		{
			name:    "nil client",
			logger:  logger,
			client:  nil,
			wantNil: false, // Should still create handler, but client will be nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prHandler := NewSDKPRListHandler(tt.logger, tt.client)
			if (prHandler == nil) != tt.wantNil {
				t.Errorf("NewSDKPRListHandler() nil = %v, want %v", prHandler == nil, tt.wantNil)
			}

			repoHandler := NewSDKRepositoryHandler(tt.logger, tt.client)
			if (repoHandler == nil) != tt.wantNil {
				t.Errorf("NewSDKRepositoryHandler() nil = %v, want %v", repoHandler == nil, tt.wantNil)
			}

			issueHandler := NewSDKIssueListHandler(tt.logger, tt.client)
			if (issueHandler == nil) != tt.wantNil {
				t.Errorf("NewSDKIssueListHandler() nil = %v, want %v", issueHandler == nil, tt.wantNil)
			}
		})
	}
}
