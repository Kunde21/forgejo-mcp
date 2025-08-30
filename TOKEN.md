# Authentication Requirements

## Token Source

Environment Variable: FORGEJO_TOKEN
Must be masked in any logs or errors.

## Validation Method

Token validation will happen on the first tool execution.
Errors will be returned to the client with a clear error message.

## Timeout

Calls to the Gitea server should time out if it exceeds 5 seconds
