package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	v "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kunde21/forgejo-mcp/remote"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NotificationList represents a collection of notifications.
// This struct is used as the result data for the notification_list tool.
type NotificationList struct {
	Notifications []remote.Notification `json:"notifications"`
	Total         int                   `json:"total"`
	Limit         int                   `json:"limit"`
	Offset        int                   `json:"offset"`
}

// NotificationListArgs represents the arguments for listing notifications
type NotificationListArgs struct {
	Repository string `json:"repository,omitzero"` // Repository path in "owner/repo" format
	Directory  string `json:"directory,omitzero"`  // Local directory path containing a git repository for automatic resolution
	Status     string `json:"status,omitzero"`     // Filter by status: "read", "unread", or "all"
	Limit      int    `json:"limit,omitzero"`      // Pagination limit (1-100, default 15)
	Offset     int    `json:"offset,omitzero"`     // Pagination offset (default 0)
}

// handleNotificationList handles the "notification_list" tool request.
// It retrieves notifications from a specified Forgejo/Gitea repository with optional filtering and pagination.
//
// Parameters:
//   - repository: The repository path in "owner/repo" format
//   - directory: Local directory path containing a git repository for automatic resolution
//   - status: Filter by notification status ("read", "unread", or "all", default "unread")
//   - limit: Maximum number of notifications to return (1-100, default 15)
//   - offset: Number of notifications to skip for pagination (default 0)
//
// Note: At least one of repository or directory must be provided. If both are provided,
// directory takes precedence for automatic repository resolution.
func (s *Server) handleNotificationList(ctx context.Context, request *mcp.CallToolRequest, args NotificationListArgs) (*mcp.CallToolResult, *NotificationList, error) {
	// Set defaults
	if args.Limit == 0 {
		args.Limit = 15
	}
	if args.Status == "" {
		args.Status = "unread" // Default to unread notifications
	}

	// Validate input arguments using ozzo-validation
	if err := v.ValidateStruct(&args,
		v.Field(&args.Repository, v.When(args.Directory == "",
			v.Required.Error("at least one of directory or repository must be provided"),
			v.Match(repoReg).Error("repository must be in format 'owner/repo'"),
		)),
		v.Field(&args.Directory, v.When(args.Repository == "",
			v.Required.Error("at least one of directory or repository must be provided"),
			v.By(func(any) error {
				if !filepath.IsAbs(args.Directory) {
					return v.NewError("abs_dir", "directory must be an absolute path")
				}
				stat, err := os.Stat(args.Directory)
				if err != nil {
					return v.NewError("abs_dir", "invalid directory")
				}
				if !stat.IsDir() {
					return v.NewError("abs_dir", "does not exist")
				}
				return nil
			}),
		)),
		v.Field(&args.Status, v.In("read", "unread", "all").Error("status must be 'read', 'unread', or 'all'")),
		v.Field(&args.Limit, v.Min(1), v.Max(100)),
		v.Field(&args.Offset, v.Min(0)),
	); err != nil {
		return TextErrorf("Invalid request: %v", err), nil, nil
	}

	repository := args.Repository
	if args.Directory != "" {
		// Resolve directory to repository (takes precedence if both provided)
		resolution, err := s.repositoryResolver.ResolveRepository(args.Directory)
		if err != nil {
			return TextErrorf("Failed to resolve directory: %v", err), nil, nil
		}
		repository = resolution.Repository
	}

	// Get remote client
	client, err := s.getRemoteClient()
	if err != nil {
		return TextErrorf("Failed to get remote client: %v", err), nil, nil
	}

	// List notifications
	notificationList, err := client.ListNotifications(ctx, repository, args.Status, args.Limit, args.Offset)
	if err != nil {
		return TextErrorf("Failed to list notifications: %v", err), nil, nil
	}

	// Format response
	responseText := fmt.Sprintf("Found %d %s notifications", len(notificationList.Notifications), args.Status)
	return TextResult(responseText), &NotificationList{
		Notifications: notificationList.Notifications,
		Total:         notificationList.Total,
		Limit:         notificationList.Limit,
		Offset:        notificationList.Offset,
	}, nil
}

// getRemoteClient returns the remote client instance
func (s *Server) getRemoteClient() (remote.ClientInterface, error) {
	if s.remote == nil {
		return nil, fmt.Errorf("remote client not initialized")
	}
	return s.remote, nil
}
