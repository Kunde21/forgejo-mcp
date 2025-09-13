# Spec Tasks

## Tasks

- [x] 1. Remove validation functions from remote/gitea/service.go
  - [x] 1.1 Write tests to verify current validation behavior before removal
  - [x] 1.2 Delete validateRepository() function (lines 156-168)
  - [x] 1.3 Delete validatePagination() function (lines 171-179)
  - [x] 1.4 Delete validateIssueNumber() function (lines 182-187)
  - [x] 1.5 Delete validatePullRequestNumber() function (lines 190-195)
  - [x] 1.6 Delete validateCommentContent() function (lines 198-207)
  - [x] 1.7 Delete validateCommentID() function (lines 210-215)
  - [x] 1.8 Delete validatePullRequestOptions() function (lines 218-230)
  - [x] 1.9 Delete validatePullRequestState() function (lines 233-245)
  - [x] 1.10 Verify all tests pass after function removal

- [x] 2. Remove validation calls from service methods in remote/gitea/service.go
  - [x] 2.1 Write tests for service methods without validation
  - [x] 2.2 Remove validation calls from ListIssues method
  - [x] 2.3 Remove validation calls from CreateIssueComment method
  - [x] 2.4 Remove validation calls from ListIssueComments method
  - [x] 2.5 Remove validation calls from EditIssueComment method
  - [x] 2.6 Remove validation calls from ListPullRequests method
  - [x] 2.7 Remove validation calls from ListPullRequestComments method
  - [x] 2.8 Verify CreatePullRequestComment and EditPullRequestComment methods remain unchanged (no validation to remove)
  - [x] 2.9 Verify all tests pass after validation call removal

- [x] 3. Remove validation tags from interface structs in remote/gitea/interface.go
  - [x] 3.1 Write tests to verify struct serialization without validation tags
  - [x] 3.2 Remove validation tags from ListIssueCommentsArgs struct
  - [x] 3.3 Remove validation tags from EditIssueCommentArgs struct
  - [x] 3.4 Remove validation tags from ListPullRequestsOptions struct
  - [x] 3.5 Remove validation tags from ListPullRequestCommentsArgs struct
  - [x] 3.6 Remove validation tags from CreatePullRequestCommentArgs struct
  - [x] 3.7 Remove validation tags from EditPullRequestCommentArgs struct
  - [x] 3.8 Verify all tests pass after validation tag removal

- [x] 4. Verify server layer validation remains intact
  - [x] 4.1 Confirm server/common.go unchanged (repoReg and helper functions preserved)
  - [x] 4.2 Verify all server handlers maintain inline validation patterns
  - [x] 4.3 Test that error messages remain consistent and user-friendly
  - [x] 4.4 Run full test suite to ensure no functional regression
  - [x] 4.5 Verify validation deduplication is complete and working correctly
