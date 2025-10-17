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
		Status: sdkStatus,
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
		notification.Type = strings.ToLower(string(thread.Subject.Type))

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
