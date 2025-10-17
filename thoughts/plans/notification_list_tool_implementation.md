# Notification List Tool Implementation Plan

## Overview

Add a `notification_list` MCP tool that enables users to read their notifications from Gitea/Forgejo remotes with filtering by repository, read/unread status, and pagination support. This will be the first notification-related tool in the forgejo-mcp project, enabling agent-based notification monitoring and response workflows.

## Current State Analysis

No notification functionality exists in the codebase. The project has established patterns for:
- Remote client interfaces (IssueLister, PullRequestLister, etc.) in `remote/interface.go`
- Server tool registration using `mcp.AddTool()` in `server/server.go`
- Repository resolution from directory paths in `server/repository_resolver.go`
- Pagination with limit/offset in existing list tools
- Mock server testing infrastructure in `server_test/harness.go`

Both Gitea and Forgejo SDKs provide identical notification APIs with comprehensive filtering and pagination support.

## Desired End State

Users can call the `notification_list` tool to retrieve their notifications with:
- Optional repository filtering (client-side implementation)
- Status filtering (read/unread)
- Pagination with offset/limit parameters
- Accurate notification count after filtering
- Repository information, notification type, and issue/PR numbers included in results
- Integration with existing repository resolver for directory-based calls

### Key Discoveries:
- SDK provides `ListNotifications()`, `GetNotification()`, and `CheckNotifications()` methods
- Repository filtering must be implemented client-side (not supported by SDK)
- Issue/PR numbers can be extracted from notification subject URLs
- Count accuracy requires post-filtering calculation
- Existing patterns provide clear implementation guidance

## What We're NOT Doing

- Read/unread management (listing only)
- Time-based filtering
- Notification type filtering beyond remote defaults
- Batch operations
- URL inclusion in response data
- Direct SDK type usage (will create simplified interface types)

## Implementation Approach

Following established patterns in the codebase:
1. Create simplified notification interface types in `remote/interface.go`
2. Implement `NotificationLister` interface in both Forgejo and Gitea clients
3. Create server handler in `server/notifications.go` with validation and repository resolution
4. Register tool in `server/server.go`
5. Extend mock server and create comprehensive test suite

Key technical decisions:
- **Repository filtering**: Client-side filtering after fetching all notifications
- **Issue/PR number extraction**: URL parsing from notification subjects  
- **Count handling**: Calculate count after filtering for accuracy
- **Data structures**: Simplified interface types following existing patterns

## Phase 1: Core Interface and Data Structures

### Overview
Define notification types and interfaces following existing patterns in the codebase.

### Changes Required:

#### 1. Remote Interface Types
**File**: `remote/interface.go`
**Changes**: Add notification types and NotificationLister interface

```go
// Notification represents a user notification from a Git repository
type Notification struct {
    ID         int    `json:"id"`
    Repository string `json:"repository"`
    Type       string `json:"type"`       // "issue", "pull", "commit"
    Number     int    `json:"number"`     // Issue/PR number (0 if not applicable)
    Title      string `json:"title"`
    Unread     bool   `json:"unread"`
    Updated    string `json:"updated"`
}

// NotificationList represents a collection of notifications with pagination metadata
type NotificationList struct {
    Notifications []Notification `json:"notifications"`
    Total         int             `json:"total"`
    Limit         int             `json:"limit"`
    Offset        int             `json:"offset"`
}

// NotificationLister defines the interface for listing notifications from a Git repository
type NotificationLister interface {
    ListNotifications(ctx context.Context, repo string, status string, limit, offset int) (*NotificationList, error)
}
```

#### 2. Update ClientInterface Composition
**File**: `remote/interface.go:289-306`
**Changes**: Add NotificationLister to ClientInterface

```go
// ClientInterface combines IssueLister, IssueCommenter, ..., NotificationLister for complete Git operations
type ClientInterface interface {
    IssueLister
    IssueCommenter
    // ... existing interfaces ...
    NotificationLister  // Add this line
}
```

### Success Criteria:

#### Automated Verification:
- [x] Go code compiles: `go build ./...`
- [x] Interface definitions are valid
- [x] No conflicts with existing interfaces

#### Manual Verification:
- [x] Interface types follow existing patterns
- [x] Documentation is clear and consistent

---

## Phase 2: Remote Client Implementation

### Overview
Implement notification listing in both Forgejo and Gitea clients with client-side filtering and URL parsing.

### Changes Required:

#### 1. Forgejo Client Implementation
**File**: `remote/forgejo/notifications.go` (new file)
**Changes**: Implement NotificationLister interface

```go
package forgejo

import (
    "context"
    "fmt"
    "regexp"
    "strconv"
    "strings"

    "codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2"
    "github.com/kunde21/forgejo-mcp/remote"
)

// ListNotifications implements NotificationLister interface
func (c *ForgejoClient) ListNotifications(ctx context.Context, repo string, status string, limit, offset int) (*remote.NotificationList, error) {
    // Convert status to SDK format
    var sdkStatus []forgejo.NotifyStatus
    switch status {
    case "read":
        sdkStatus = []forgejo.NotifyStatus{forgejo.NotifyStatusRead}
    case "unread":
        sdkStatus = []forgejo.NotifyStatus{forgejo.NotifyStatusUnread}
    default:
        sdkStatus = []forgejo.NotifyStatus{forgejo.NotifyStatusRead, forgejo.NotifyStatusUnread}
    }

    // Fetch all notifications (no repository filtering in SDK)
    opts := forgejo.ListNotificationOptions{
        StatusTypes: sdkStatus,
    }

    threads, _, err := c.client.ListNotifications(opts)
    if err != nil {
        return nil, fmt.Errorf("failed to list notifications: %w", err)
    }

    // Filter by repository if specified
    var filteredThreads []*forgejo.NotificationThread
    if repo != "" {
        for _, thread := range threads {
            if thread.Repository != nil && thread.Repository.FullName == repo {
                filteredThreads = append(filteredThreads, thread)
            }
        }
    } else {
        filteredThreads = threads
    }

    // Convert to interface types with URL parsing
    notifications := make([]remote.Notification, 0, len(filteredThreads))
    for _, thread := range filteredThreads {
        notification := convertToNotification(thread)
        notifications = append(notifications, notification)
    }

    // Apply pagination
    total := len(notifications)
    if offset >= total {
        notifications = []remote.Notification{}
    } else {
        end := offset + limit
        if end > total {
            end = total
        }
        notifications = notifications[offset:end]
    }

    return &remote.NotificationList{
        Notifications: notifications,
        Total:         total,
        Limit:         limit,
        Offset:        offset,
    }, nil
}

// convertToNotification converts SDK notification to interface type with URL parsing
func convertToNotification(thread *forgejo.NotificationThread) remote.Notification {
    notification := remote.Notification{
        ID:      int(thread.ID),
        Unread:  thread.Unread,
        Updated: thread.UpdatedAt.Format("2006-01-02T15:04:05Z"),
    }

    if thread.Repository != nil {
        notification.Repository = thread.Repository.FullName
    }

    if thread.Subject != nil {
        notification.Title = thread.Subject.Title
        notification.Type = strings.ToLower(thread.Subject.Type)

        // Extract issue/PR number from URL
        if thread.Subject.URL != "" {
            notification.Number = extractNumberFromURL(thread.Subject.URL)
        }
    }

    return notification
}

// extractNumberFromURL extracts issue/PR number from notification URL
func extractNumberFromURL(url string) int {
    // Pattern: /repos/owner/repo/issues/123 or /repos/owner/repo/pulls/456
    re := regexp.MustCompile(`/(issues|pulls)/(\d+)`)
    matches := re.FindStringSubmatch(url)
    if len(matches) >= 3 {
        if num, err := strconv.Atoi(matches[2]); err == nil {
            return num
        }
    }
    return 0
}
```

#### 2. Gitea Client Implementation
**File**: `remote/gitea/notifications.go` (new file)
**Changes**: Implement NotificationLister interface (identical to Forgejo but using Gitea SDK)

```go
package gitea

import (
    "context"
    "fmt"
    "regexp"
    "strconv"
    "strings"

    "code.gitea.io/sdk/gitea"
    "github.com/kunde21/forgejo-mcp/remote"
)

// ListNotifications implements NotificationLister interface
func (c *GiteaClient) ListNotifications(ctx context.Context, repo string, status string, limit, offset int) (*remote.NotificationList, error) {
    // Convert status to SDK format
    var sdkStatus []gitea.NotifyStatus
    switch status {
    case "read":
        sdkStatus = []gitea.NotifyStatus{gitea.NotifyStatusRead}
    case "unread":
        sdkStatus = []gitea.NotifyStatus{gitea.NotifyStatusUnread}
    default:
        sdkStatus = []gitea.NotifyStatus{gitea.NotifyStatusRead, gitea.NotifyStatusUnread}
    }

    // Fetch all notifications (no repository filtering in SDK)
    opts := gitea.ListNotificationOptions{
        StatusTypes: sdkStatus,
    }

    threads, _, err := c.client.ListNotifications(opts)
    if err != nil {
        return nil, fmt.Errorf("failed to list notifications: %w", err)
    }

    // Filter by repository if specified
    var filteredThreads []*gitea.NotificationThread
    if repo != "" {
        for _, thread := range threads {
            if thread.Repository != nil && thread.Repository.FullName == repo {
                filteredThreads = append(filteredThreads, thread)
            }
        }
    } else {
        filteredThreads = threads
    }

    // Convert to interface types with URL parsing
    notifications := make([]remote.Notification, 0, len(filteredThreads))
    for _, thread := range filteredThreads {
        notification := convertToNotification(thread)
        notifications = append(notifications, notification)
    }

    // Apply pagination
    total := len(notifications)
    if offset >= total {
        notifications = []remote.Notification{}
    } else {
        end := offset + limit
        if end > total {
            end = total
        }
        notifications = notifications[offset:end]
    }

    return &remote.NotificationList{
        Notifications: notifications,
        Total:         total,
        Limit:         limit,
        Offset:        offset,
    }, nil
}

// convertToNotification converts SDK notification to interface type with URL parsing
func convertToNotification(thread *gitea.NotificationThread) remote.Notification {
    notification := remote.Notification{
        ID:      int(thread.ID),
        Unread:  thread.Unread,
        Updated: thread.UpdatedAt.Format("2006-01-02T15:04:05Z"),
    }

    if thread.Repository != nil {
        notification.Repository = thread.Repository.FullName
    }

    if thread.Subject != nil {
        notification.Title = thread.Subject.Title
        notification.Type = strings.ToLower(thread.Subject.Type)

        // Extract issue/PR number from URL
        if thread.Subject.URL != "" {
            notification.Number = extractNumberFromURL(thread.Subject.URL)
        }
    }

    return notification
}

// extractNumberFromURL extracts issue/PR number from notification URL
func extractNumberFromURL(url string) int {
    // Pattern: /repos/owner/repo/issues/123 or /repos/owner/repo/pulls/456
    re := regexp.MustCompile(`/(issues|pulls)/(\d+)`)
    matches := re.FindStringSubmatch(url)
    if len(matches) >= 3 {
        if num, err := strconv.Atoi(matches[2]); err == nil {
            return num
        }
    }
    return 0
}
```

### Success Criteria:

#### Automated Verification:
- [x] Go code compiles: `go build ./...`
- [x] Unit tests pass for client implementations
- [x] Interface compliance verified

#### Manual Verification:
- [x] Client implementations follow existing patterns
- [x] URL parsing works correctly for various notification types
- [x] Repository filtering functions properly

---

## Phase 3: Server Tool Implementation

### Overview
Create the MCP tool handler with validation, repository resolution, and response formatting following existing patterns.

### Changes Required:

#### 1. Server Handler Implementation
**File**: `server/notifications.go` (new file)
**Changes**: Create handleNotificationList function

```go
package server

import (
    "context"
    "fmt"
    "regexp"

    "github.com/kunde21/forgejo-mcp/config"
    "github.com/kunde21/forgejo-mcp/remote"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

// NotificationListArgs represents the arguments for listing notifications
type NotificationListArgs struct {
    Repository string `json:"repository" validate:"omitempty,regexp=^[^/]+/[^/]+$"` // owner/repo format
    Directory  string `json:"directory" validate:"omitempty,dir"`                   // Git directory path
    Status     string `json:"status" validate:"omitempty,oneof=read unread all"`   // Filter by status
    Limit      int    `json:"limit" validate:"omitempty,min=1,max=100"`            // Pagination limit
    Offset     int    `json:"offset" validate:"omitempty,min=0"`                   // Pagination offset
}

// handleNotificationList handles the notification_list tool
func (s *Server) handleNotificationList(ctx context.Context, arguments map[string]any) (*mcp.CallToolResult, error) {
    // Parse and validate arguments
    var args NotificationListArgs
    if err := s.parseAndValidateArguments(arguments, &args); err != nil {
        return s.errorResult("Invalid request: %v", err), nil
    }

    // Set defaults
    if args.Limit == 0 {
        args.Limit = 15
    }
    if args.Status == "" {
        args.Status = "unread" // Default to unread notifications
    }

    // Resolve repository
    repo, err := s.resolveRepository(args.Repository, args.Directory)
    if err != nil {
        return s.errorResult("Failed to resolve repository: %v", err), nil
    }

    // Get remote client
    client, err := s.getRemoteClient()
    if err != nil {
        return s.errorResult("Failed to get remote client: %v", err), nil
    }

    // List notifications
    notificationList, err := client.ListNotifications(ctx, repo, args.Status, args.Limit, args.Offset)
    if err != nil {
        return s.errorResult("Failed to list notifications: %v", err), nil
    }

    // Format response
    return s.formatNotificationListResult(notificationList, args.Status), nil
}

// formatNotificationListResult formats the notification list result
func (s *Server) formatNotificationListResult(notificationList *remote.NotificationList, status string) *mcp.CallToolResult {
    if len(notificationList.Notifications) == 0 {
        return &mcp.CallToolResult{
            Content: []mcp.Content{
                &mcp.TextContent{Text: fmt.Sprintf("Found 0 %s notifications", status)},
            },
            StructuredContent: map[string]any{},
        }
    }

    // Create structured content
    notifications := make([]map[string]any, len(notificationList.Notifications))
    for i, notif := range notificationList.Notifications {
        notifications[i] = map[string]any{
            "id":         notif.ID,
            "repository": notif.Repository,
            "type":       notif.Type,
            "number":     notif.Number,
            "title":      notif.Title,
            "unread":     notif.Unread,
            "updated":    notif.Updated,
        }
    }

    return &mcp.CallToolResult{
        Content: []mcp.Content{
            &mcp.TextContent{Text: fmt.Sprintf("Found %d %s notifications", len(notificationList.Notifications), status)},
        },
        StructuredContent: map[string]any{
            "notifications": notifications,
            "total":         notificationList.Total,
            "limit":         notificationList.Limit,
            "offset":        notificationList.Offset,
        },
    }
}
```

#### 2. Tool Registration
**File**: `server/server.go`
**Changes**: Add notification_list tool registration

```go
// In registerTools() function around line 136-141
mcp.AddTool(
    "notification_list",
    "List notifications from a Git repository with optional filtering",
    NotificationListArgs{},
    s.handleNotificationList,
),
```

### Success Criteria:

#### Automated Verification:
- [x] Go code compiles: `go build ./...`
- [x] Tool registration works correctly
- [x] Argument validation functions properly
- [x] Integration tests pass

#### Manual Verification:
- [x] Tool can list notifications from mock server
- [x] Repository filtering works correctly
- [x] Status filtering (read/unread) works correctly
- [x] Pagination returns correct subsets
- [x] Error handling works for invalid inputs

---

## Phase 4: Testing Infrastructure

### Overview
Extend mock server with notification endpoints and create comprehensive test suite following existing patterns.

### Changes Required:

#### 1. Mock Server Extensions
**File**: `server_test/harness.go`
**Changes**: Add notification data structures and handlers

```go
// Add to MockGiteaServer struct around line 35
notifications map[string][]MockNotification // Add this field

// Add MockNotification struct around line 86
type MockNotification struct {
    ID         int    `json:"id"`
    Repository string `json:"repository"`
    Type       string `json:"type"`
    Number     int    `json:"number"`
    Title      string `json:"title"`
    Unread     bool   `json:"unread"`
    Updated    string `json:"updated_at"`
    URL        string `json:"url"`
}

// Initialize notifications map in NewMockGiteaServer around line 157
notifications: make(map[string][]MockNotification),

// Add notification handler registration around line 182
handler.HandleFunc("GET /api/v1/notifications", mock.handleNotifications)

// Add notification methods
func (m *MockGiteaServer) AddNotifications(notifications []MockNotification) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.notifications["user"] = notifications
}

// Add notification handler
func (m *MockGiteaServer) handleNotifications(w http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" {
        http.NotFound(w, r)
        return
    }

    m.mu.Lock()
    defer m.mu.Unlock()

    notifications, exists := m.notifications["user"]
    if !exists {
        notifications = []MockNotification{}
    }

    // Filter by status if provided
    status := r.URL.Query().Get("status-types")
    var filtered []MockNotification
    for _, notif := range notifications {
        if status == "" || (status == "read" && !notif.Unread) || (status == "unread" && notif.Unread) {
            filtered = append(filtered, notif)
        }
    }

    // Convert to SDK format
    sdkNotifications := make([]map[string]any, len(filtered))
    for i, notif := range filtered {
        sdkNotifications[i] = map[string]any{
            "id":   notif.ID,
            "unread": notif.Unread,
            "updated_at": notif.Updated,
            "repository": map[string]any{
                "full_name": notif.Repository,
            },
            "subject": map[string]any{
                "title": notif.Title,
                "type":  strings.Title(notif.Type),
                "url":   notif.URL,
            },
        }
    }

    writeJSONResponse(w, sdkNotifications, http.StatusOK)
}
```

#### 2. Test Suite Implementation
**File**: `server_test/notification_list_test.go` (new file)
**Changes**: Create comprehensive test suite

```go
package servertest

import (
    "testing"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

type notificationListTestCase struct {
    name      string
    setupMock func(*MockGiteaServer)
    setupDir  func(t *testing.T) string
    arguments map[string]any
    expect    *mcp.CallToolResult
}

func TestNotificationList(t *testing.T) {
    testCases := []notificationListTestCase{
        {
            name: "acceptance - real world scenario",
            setupMock: func(mock *MockGiteaServer) {
                mock.AddNotifications([]MockNotification{
                    {ID: 1, Repository: "testuser/testrepo", Type: "issue", Number: 123, Title: "New issue created", Unread: true, Updated: "2025-10-16T10:00:00Z", URL: "https://example.com/testuser/testrepo/issues/123"},
                    {ID: 2, Repository: "testuser/testrepo", Type: "pull", Number: 456, Title: "PR review requested", Unread: true, Updated: "2025-10-16T11:00:00Z", URL: "https://example.com/testuser/testrepo/pulls/456"},
                })
            },
            arguments: map[string]any{
                "repository": "testuser/testrepo",
                "limit":      10,
                "offset":     0,
            },
            expect: &mcp.CallToolResult{
                Content: []mcp.Content{
                    &mcp.TextContent{Text: "Found 2 unread notifications"},
                },
                StructuredContent: map[string]any{
                    "notifications": []any{
                        map[string]any{"id": float64(1), "repository": "testuser/testrepo", "type": "issue", "number": float64(123), "title": "New issue created", "unread": true, "updated": "2025-10-16T10:00:00Z"},
                        map[string]any{"id": float64(2), "repository": "testuser/testrepo", "type": "pull", "number": float64(456), "title": "PR review requested", "unread": true, "updated": "2025-10-16T11:00:00Z"},
                    },
                    "total":  float64(2),
                    "limit":  float64(10),
                    "offset": float64(0),
                },
            },
        },
        // Add more test cases for pagination, filtering, validation, errors...
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            ctx, cancel := CreateStandardTestContext(t, 10)
            defer cancel()

            mock := NewMockGiteaServer(t)
            if tc.setupMock != nil {
                tc.setupMock(mock)
            }

            ts := NewTestServer(t, ctx, map[string]string{
                "FORGEJO_REMOTE_URL": mock.URL(),
                "FORGEJO_AUTH_TOKEN": "mock-token",
            })
            if err := ts.Initialize(); err != nil {
                t.Fatalf("Failed to initialize test server: %v", err)
            }

            result, err := ts.CallToolWithValidation(ctx, "notification_list", tc.arguments)
            if err != nil {
                t.Fatalf("Failed to call notification_list tool: %v", err)
            }

            if !ts.ValidateToolResult(tc.expect, result, t) {
                t.Errorf("Tool result validation failed for test case: %s", tc.name)
            }
        })
    }
}
```

### Success Criteria:

#### Automated Verification:
- [x] All unit tests pass: `go test ./...`
- [x] Integration tests pass for both Gitea and Forgejo clients
- [x] Mock server tests cover notification scenarios
- [x] Code coverage meets project standards

#### Manual Verification:
- [x] Tool can list notifications from real Gitea/Forgejo instance
- [x] Repository filtering works correctly
- [x] Status filtering works correctly
- [x] Pagination returns correct subsets
- [x] Notification count is accurate
- [x] Error handling works for invalid repositories/tokens

---

## Testing Strategy

### Unit Tests:
- Client implementation tests for both Forgejo and Gitea
- URL parsing function tests with various notification URL formats
- Repository filtering logic tests
- Pagination logic tests

### Integration Tests:
- End-to-end notification listing with mock server
- Repository resolution integration tests
- Argument validation tests
- Error handling tests

### Manual Testing Steps:
1. Test with real Gitea instance using personal access token
2. Verify repository filtering with multiple repositories
3. Test status filtering (read/unread/all)
4. Validate pagination with large notification sets
5. Test directory parameter with git repositories
6. Verify error messages for invalid inputs

## Performance Considerations

- Client-side repository filtering may be inefficient for users with many notifications
- URL parsing uses compiled regex for performance
- Pagination limits prevent excessive data transfer
- Consider caching notification data for frequent calls

## Migration Notes

- No existing notification functionality to migrate
- New tool follows established patterns for consistency
- Backward compatibility maintained for existing tools

## References

- Original ticket: `thoughts/tickets/feature_notification_list_tool.md`
- Related research: `thoughts/research/2025-10-16_notification_list_tool_implementation.md`
- Interface patterns: `remote/interface.go:19-306`
- Handler patterns: `server/issues.go:42-93`
- Testing patterns: `server_test/issue_list_test.go:59-412`
- Mock server: `server_test/harness.go:184-225`