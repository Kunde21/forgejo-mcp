# Spec Tasks

## Tasks

- [ ] 1. Remove validation functions from remote/gitea/service.go
  - [ ] 1.1 Write tests to verify current validation behavior before removal
  - [ ] 1.2 Delete validateRepository() function (lines 156-168)
  - [ ] 1.3 Delete validatePagination() function (lines 171-179)
  - [ ] 1.4 Delete validateIssueNumber() function (lines 182-187)
  - [ ] 1.5 Delete validatePullRequestNumber() function (lines 190-195)
  - [ ] 1.6 Delete validateCommentContent() function (lines 198-207)
  - [ ] 1.7 Delete validateCommentID() function (lines 210-215)
  - [ ] 1.8 Delete validatePullRequestOptions() function (lines 218-230)
  - [ ] 1.9 Delete validatePullRequestState() function (lines 233-245)
  - [ ] 1.10 Verify all tests pass after function removal

- [ ] 2. Remove validation calls from service methods in remote/gitea/service.go
  - [ ] 2.1 Write tests for service methods without validation
  - [ ] 2.2 Remove validation calls from ListIssues method
  - [ ] 2.3 Remove validation calls from CreateIssueComment method
  - [ ] 2.4 Remove validation calls from ListIssueComments method
  - [ ] 2.5 Remove validation calls from EditIssueComment method
  - [ ] 2.6 Remove validation calls from ListPullRequests method
  - [ ] 2.7 Remove validation calls from ListPullRequestComments method
  - [ ] 2.8 Verify CreatePullRequestComment and EditPullRequestComment methods remain unchanged (no validation to remove)
  - [ ] 2.9 Verify all tests pass after validation call removal

- [ ] 3. Remove validation tags from interface structs in remote/gitea/interface.go
  - [ ] 3.1 Write tests to verify struct serialization without validation tags
  - [ ] 3.2 Remove validation tags from ListIssueCommentsArgs struct
  - [ ] 3.3 Remove validation tags from EditIssueCommentArgs struct
  - [ ] 3.4 Remove validation tags from ListPullRequestsOptions struct
  - [ ] 3.5 Remove validation tags from ListPullRequestCommentsArgs struct
  - [ ] 3.6 Remove validation tags from CreatePullRequestCommentArgs struct
  - [ ] 3.7 Remove validation tags from EditPullRequestCommentArgs struct
  - [ ] 3.8 Verify all tests pass after validation tag removal

- [ ] 4. Verify server layer validation remains intact
  - [ ] 4.1 Confirm server/common.go unchanged (repoReg and helper functions preserved)
  - [ ] 4.2 Verify all server handlers maintain inline validation patterns
  - [ ] 4.3 Test that error messages remain consistent and user-friendly
  - [ ] 4.4 Run full test suite to ensure no functional regression
  - [ ] 4.5 Verify validation deduplication is complete and working correctly