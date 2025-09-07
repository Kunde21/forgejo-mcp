Plan the addition of a "list issues" tool that will connect to a gitea instance to list issues for a given repository.

Arguments will include
- Repository (format "owner/repository")
- Limit for pagination (default 15)
- Offset for pagination (default 0)

Response will include
- PR number
- PR Title
- PR Status (WIP, open, closed, merged)

Configuration additions
- Add REMOTE_URL to identify the Gitea or Forgejo instance
- Add AUTH_TOKEN to authenticate the calls with the Gitea SDK

Package additions
- Add `remote/gitea` directory to implement the handler functionality using the gitea sdk 
- Use dependency injection with a receiver interface defined in the server package, allow loading implementations using gitea or forgejo sdk 

Additions to the test harness:
- Use `httptest.Server` to serve a mock Gitea API and configure the mcp client to use the server URL for test scenarios.
- Add acceptance tests for the tool calls using the test harness.
- Once the acceptance tests validate the tool calls are working correctly and connecting to the api server, those tests will be maintained to identify regressions.  Remove tests for implementation details.
