# Product Mission

> Last Updated: 2025-08-26
> Version: 1.0.0

## Pitch

Forgejo MCP is a Model Context Protocol server that enables AI agents to interact with Forgejo repositories through standardized CLI tools. It provides a bridge between AI workflows and Forgejo's Git repository management features.

## Users

- AI agents using the Model Context Protocol
- Developers working with Forgejo repositories
- Automation workflows that need to interact with Git repositories

## The Problem

AI agents need a standardized way to interact with Forgejo repositories for common development tasks like managing pull requests and issues. Currently, there's no standardized interface for AI agents to perform these operations directly on Forgejo repositories.

## Differentiators

- Built specifically for the Model Context Protocol standard
- Leverages existing `tea` CLI tool functionality from Gitea
- Works with any Forgejo repository that has a remote configured
- Provides a clean interface for AI agents to perform repository management tasks

## Key Features

- Pull Request Management (list, comment, review)
- Issue Management (list, create, close, comment)
- CLI-based interaction model
- External authentication model
- Repository context awareness