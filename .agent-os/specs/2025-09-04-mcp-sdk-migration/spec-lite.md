# Spec Summary (Lite)

Replace the current hybrid MCP implementation with the official MCP SDK to ensure full protocol compliance and reduce maintenance overhead. This migration will eliminate ~2000+ lines of custom MCP server code while preserving all existing Forgejo integration functionality, enabling easier maintenance and automatic protocol updates through the official SDK.