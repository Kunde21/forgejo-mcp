# PR Creation Tool Documentation

## Overview

The `pr_create` tool enables users to create pull requests in Forgejo and Gitea repositories through the MCP (Model Context Protocol) interface. It provides comprehensive functionality including automatic repository resolution, fork detection, template loading, and advanced conflict detection.

## Features

### ✅ Core Functionality
- **Pull Request Creation**: Create PRs with title, description, and metadata
- **Branch Validation**: Automatic detection and validation of source/target branches
- **Conflict Detection**: Advanced conflict analysis with detailed file reporting
- **Draft Support**: Create draft PRs for work-in-progress changes

### ✅ Advanced Features
- **Fork Detection**: Automatically detects fork relationships and targets original repositories
- **Template Loading**: Loads and merges PR templates from repository (`.gitea/`, `.github/`, etc.)
- **Directory Parameter**: Automatic repository resolution from local git directories
- **Enhanced Error Messages**: Detailed, actionable error messages with suggested solutions

### ✅ Git Integration
- **Auto-detection**: Automatically detects current branch when not specified
- **Conflict Analysis**: Detailed conflict reporting with file lists and resolution suggestions
- **Branch Status**: Checks if branches are behind, ahead, or have diverged

## Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `repository` | string | No* | Repository path in "owner/repo" format |
| `directory` | string | No* | Local directory path containing a git repository |
| `title` | string | Yes | Pull request title (1-255 characters) |
| `body` | string | No | Pull request description (1-65535 characters) |
| `head` | string | No | Source branch (auto-detected if not provided) |
| `base` | string | No | Target branch (defaults to "main") |
| `draft` | boolean | No | Create as draft PR (default: false) |
| `assignee` | string | No | Single reviewer to assign |

*At least one of `repository` or `directory` must be provided. If both are provided, `directory` takes precedence.

## Usage Examples

### Basic PR Creation

```json
{
  "repository": "owner/repo",
  "title": "Add new feature",
  "body": "This PR adds a new feature to improve user experience.",
  "head": "feature-branch",
  "base": "main"
}
```

### Directory-based Creation (Recommended)

```json
{
  "directory": "/path/to/your/project",
  "title": "Fix authentication bug"
}
```

### Draft PR Creation

```json
{
  "repository": "owner/repo",
  "title": "WIP: New dashboard design",
  "draft": true,
  "head": "dashboard-redesign"
}
```

### With Assignee

```json
{
  "repository": "owner/repo",
  "title": "Update dependencies",
  "assignee": "maintainer-user"
}
```

## Advanced Features

### Fork Detection

When using the `directory` parameter, the tool automatically detects fork relationships:

```json
{
  "directory": "/path/to/forked-repo",
  "title": "Fix bug in original repo"
}
```

If the directory is a fork, the tool will:
- Detect the fork relationship
- Target the original repository for PR creation
- Include fork information in the response

### Template Loading

The tool automatically loads PR templates from the repository in this order:

1. `.gitea/PULL_REQUEST_TEMPLATE.md`
2. `.github/PULL_REQUEST_TEMPLATE.md`
3. `PULL_REQUEST_TEMPLATE.md`
4. `docs/PULL_REQUEST_TEMPLATE.md`

If no `body` is provided, the template will be used. If both template and `body` are provided, they will be merged using placeholder replacement.

### Template Placeholders

Templates can use these placeholders:
- `{{title}}` - Replaced with the PR title
- `{{description}}` - Replaced with the user-provided body
- `{{body}}` - Replaced with the user-provided body

Example template:
```markdown
## Description
{{description}}

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
```

## Conflict Detection

The tool provides detailed conflict analysis:

### Conflict Report Structure

```json
{
  "has_conflicts": true,
  "conflict_files": ["src/main.go", "README.md"],
  "total_conflicts": 3,
  "conflict_details": [
    {
      "file": "src/main.go",
      "type": "content",
      "lines": [45, 67, 89],
      "severity": "medium"
    }
  ],
  "suggested_actions": [
    "Resolve conflicts in affected files",
    "Test your changes after resolving conflicts",
    "Consider rebasing your branch"
  ]
}
```

### Conflict Severity Levels

- **Low**: 1-3 conflict lines
- **Medium**: 4-10 conflict lines  
- **High**: 10+ conflict lines

## Error Handling

The tool provides detailed error messages with actionable suggestions:

### Repository Resolution Errors

```
Failed to resolve directory '/path/to/project': Not a git repository (no .git directory found). 
Please run 'git init' in this directory or navigate to a valid git repository.
```

### Branch Validation Errors

```
Source branch 'feature-xyz' does not exist in directory '/path/to/project'. 
Available branches:
• Use 'git branch' to list local branches
• Use 'git checkout -b feature-xyz' to create it
• Use 'git fetch origin' to update remote branches
```

### Conflict Errors

```
Branch 'feature-branch' has 2 conflicts with 'main'.

Conflict Analysis:
- Total conflicts: 2
- Files affected: 1
- Conflicting files:
  • src/main.go
- Suggested actions:
  • Resolve conflicts in affected files
  • Test your changes after resolving conflicts
```

### API Errors

```
Failed to create pull request in 'owner/repo' from 'feature' to 'main': 
Authentication failed. Please check your API token is valid and has pull request permissions.
```

## Response Format

### Success Response

```
Pull request created successfully. Number: 123, Title: Add new feature, State: open, Created: 2025-10-07T12:00:00Z
Body: This PR adds a new feature to improve user experience.
Fork Information: Created from fork 'myuser' targeting original repository 'owner'
Template: Used repository PR template for description
```

### Error Response

```
Branch 'feature' has 3 conflicts with 'main'.

Conflict Analysis:
- Total conflicts: 3
- Files affected: 2
- Conflicting files:
  • src/main.go
  • README.md
- Suggested actions:
  • Single file has 3 conflicts - review and resolve manually
  • Use 'git merge-base' to find common ancestor for better context
  • Test your changes after resolving conflicts
```

## Best Practices

### 1. Use Directory Parameter

Prefer the `directory` parameter over `repository` for automatic features:
- Fork detection
- Template loading
- Branch auto-detection
- Conflict analysis

### 2. Provide Meaningful Titles

```json
{
  "title": "Fix: Authentication timeout issue"  // Good
}
```

### 3. Use Templates

Create PR templates in your repository for consistent descriptions:
- `.gitea/PULL_REQUEST_TEMPLATE.md` (Forgejo)
- `.github/PULL_REQUEST_TEMPLATE.md` (GitHub compatibility)

### 4. Handle Conflicts

When conflicts are detected:
1. Review the detailed conflict analysis
2. Resolve conflicts in your local branch
3. Test your changes
4. Retry PR creation

### 5. Use Draft PRs

For work-in-progress changes:
```json
{
  "title": "WIP: New feature implementation",
  "draft": true
}
```

## Integration Examples

### MCP Client Usage

```javascript
const result = await client.callTool("pr_create", {
  directory: "/path/to/project",
  title: "Add user authentication",
  body: "Implements OAuth2 authentication flow"
});

console.log(result.content[0].text);
```

### CLI Usage

```bash
# Using directory parameter
mcp call pr_create --directory /path/to/project --title "Fix bug"

# Using repository parameter
mcp call pr_create --repository owner/repo --title "New feature" --head feature-branch
```

## Troubleshooting

### Common Issues

1. **"Directory does not exist"**
   - Check the path is absolute
   - Verify the directory exists
   - Ensure proper permissions

2. **"Not a git repository"**
   - Run `git init` in the directory
   - Navigate to the correct git repository

3. **"No remotes configured"**
   - Add a remote: `git remote add origin <url>`
   - Check with `git remote -v`

4. **"Branch does not exist"**
   - Create the branch: `git checkout -b branch-name`
   - Fetch remotes: `git fetch origin`

5. **"Authentication failed"**
   - Check API token validity
   - Verify token has PR creation permissions
   - Ensure token isn't expired

### Debug Mode

Enable verbose logging for troubleshooting:
- Check server logs for detailed error information
- Use the enhanced error messages for guidance
- Verify git repository state with `git status`

## Security Considerations

- API tokens should be kept secure and have minimal required permissions
- Directory paths are validated to prevent path traversal
- Repository access is validated through token permissions
- Input validation prevents injection attacks

## Performance Notes

- Conflict detection uses efficient git commands
- Template loading is cached per request
- Repository resolution is optimized for common git configurations
- Large repositories may have longer conflict detection times

## Compatibility

- **Forgejo**: Full support with all features
- **Gitea**: Full support with all features  
- **Git**: Required for directory-based features
- **MCP SDK**: Compatible with v0.4.0+

## Future Enhancements

- Multiple assignee support
- Label and milestone assignment
- PR template inheritance from organization
- Advanced conflict resolution suggestions
- Integration with issue tracking