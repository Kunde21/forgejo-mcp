package main

import (
	"fmt"
	"github.com/Kunde21/forgejo-mcp/server"
)

func main() {
	// Test validateRepositoryFormat
	fmt.Println("Testing validateRepositoryFormat:")
	
	testCases := []string{
		"owner/repo",
		"",
		"ownerrepo", 
		"/repo",
		"owner/",
		"user123/repo456",
	}
	
	for _, tc := range testCases {
		valid, err := server.validateRepositoryFormat(tc)
		fmt.Printf("  %q -> Valid: %v, Error: %v\n", tc, valid, err)
	}
}
