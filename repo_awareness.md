# Repository awareness

MCP server will detect if the repository has a gitea or forgejo remote, attempt to identify the upstream remote if there are multiple. 

Once the remote has been identified, use that owner and repository name to fetch the issues and PRs from that repository.  Calls will use `ListRepoIssues` and `ListRepoPullRequests` methods of the gitea SDK.
