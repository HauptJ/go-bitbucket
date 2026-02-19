package bitbucket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDecodePullRequests_Success(t *testing.T) {
	// Valid pull request response
	input := map[string]interface{}{
		"id":          float64(123),
		"title":       "Test PR",
		"description": "Test description",
		"state":       "OPEN",
		"draft":       false,
		"created_on":  "2026-01-15T10:30:00.000000+00:00",
		"updated_on":  "2026-01-16T14:20:00.000000+00:00",
		"author": map[string]interface{}{
			"display_name": "John Doe",
			"uuid":         "{12345678-1234-1234-1234-123456789012}",
		},
		"source": map[string]interface{}{
			"branch": map[string]interface{}{
				"name": "feature-branch",
			},
			"repository": map[string]interface{}{
				"full_name": "owner/repo",
			},
			"commit": map[string]interface{}{
				"hash": "abc123",
			},
		},
		"destination": map[string]interface{}{
			"branch": map[string]interface{}{
				"name": "main",
			},
			"commit": map[string]interface{}{
				"hash": "def456",
			},
		},
		"close_source_branch": true,
	}

	result, err := decodePullRequests(input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(123), result.ID)
	assert.Equal(t, "Test PR", result.Title)
	assert.Equal(t, "Test description", result.Description)
	assert.Equal(t, "OPEN", result.State)
	assert.False(t, result.Draft)
	assert.True(t, result.CloseSourceBranch)
	assert.NotNil(t, result.CreatedOnTime)
	assert.NotNil(t, result.UpdatedOnTime)
}

func TestDecodePullRequests_ErrorType(t *testing.T) {
	// Response with error type
	input := map[string]interface{}{
		"type": "error",
		"error": map[string]interface{}{
			"message": "Pull request not found",
		},
	}

	result, err := decodePullRequests(input)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestDecodePullRequests_InvalidTime(t *testing.T) {
	// Invalid time format
	input := map[string]interface{}{
		"id":         float64(123),
		"title":      "Test PR",
		"created_on": "invalid-time-format",
		"updated_on": "2026-01-16T14:20:00.000000+00:00",
	}

	result, err := decodePullRequests(input)
	assert.Error(t, err)

	// Should still decode but may have issues with time parsing
	// The actual behavior depends on the stringToTimeHookFunc implementation
	assert.Nil(t, result)
}

func TestDecodePullRequests_MinimalData(t *testing.T) {
	// Minimal valid data
	input := map[string]interface{}{
		"id": float64(1),
	}

	result, err := decodePullRequests(input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(1), result.ID)
	assert.Equal(t, "", result.Title)
	assert.Equal(t, "", result.State)
}

func TestDecodePullRequestsList_Success(t *testing.T) {
	// Valid pull requests list response
	input := map[string]interface{}{
		"page":    float64(1),
		"pagelen": float64(10),
		"size":    float64(2),
		"values": []interface{}{
			map[string]interface{}{
				"id":         float64(1),
				"title":      "First PR",
				"state":      "OPEN",
				"draft":      false,
				"created_on": "2026-01-15T10:30:00.000000+00:00",
			},
			map[string]interface{}{
				"id":         float64(2),
				"title":      "Second PR",
				"state":      "MERGED",
				"draft":      true,
				"created_on": "2026-01-16T11:30:00.000000+00:00",
			},
		},
	}

	result, err := decodePullRequestsList(input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(1), result.Page)
	assert.Equal(t, uint(10), result.Pagelen)
	assert.Equal(t, uint(2), result.Size)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, uint(1), result.Items[0].ID)
	assert.Equal(t, "First PR", result.Items[0].Title)
	assert.Equal(t, "OPEN", result.Items[0].State)
	assert.False(t, result.Items[0].Draft)
	assert.Equal(t, uint(2), result.Items[1].ID)
	assert.Equal(t, "Second PR", result.Items[1].Title)
	assert.Equal(t, "MERGED", result.Items[1].State)
	assert.True(t, result.Items[1].Draft)
}

func TestDecodePullRequestsList_InvalidFormat(t *testing.T) {
	// Invalid format - not a map
	input := "invalid string"

	result, err := decodePullRequestsList(input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Not a valid format")
}

func TestDecodePullRequestsList_EmptyValues(t *testing.T) {
	// Empty values array
	input := map[string]interface{}{
		"page":    float64(1),
		"pagelen": float64(10),
		"size":    float64(0),
		"values":  []interface{}{},
	}

	result, err := decodePullRequestsList(input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(1), result.Page)
	assert.Equal(t, uint(10), result.Pagelen)
	assert.Equal(t, uint(0), result.Size)
	assert.Len(t, result.Items, 0)
}

func TestDecodePullRequestsList_MissingPaginationFields(t *testing.T) {
	// Missing pagination fields - should default to 0
	input := map[string]interface{}{
		"values": []interface{}{
			map[string]interface{}{
				"id":    float64(1),
				"title": "Test PR",
			},
		},
	}

	result, err := decodePullRequestsList(input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(0), result.Page)
	assert.Equal(t, uint(0), result.Pagelen)
	assert.Equal(t, uint(0), result.Size)
	assert.Len(t, result.Items, 1)
}

func TestDecodePullRequestsList_WithErrorInItem(t *testing.T) {
	// One valid PR, one with error type (should be skipped)
	input := map[string]interface{}{
		"page":    float64(1),
		"pagelen": float64(10),
		"size":    float64(2),
		"values": []interface{}{
			map[string]interface{}{
				"id":    float64(1),
				"title": "Valid PR",
			},
			map[string]interface{}{
				"type": "error",
				"error": map[string]interface{}{
					"message": "Error in PR",
				},
			},
		},
	}

	result, err := decodePullRequestsList(input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Only the valid PR should be in the list
	assert.Len(t, result.Items, 1)
	assert.Equal(t, uint(1), result.Items[0].ID)
}

func TestDecodePullRequestsComments_Success(t *testing.T) {
	// Valid comment response
	input := map[string]interface{}{
		"id": float64(756186348),
		"content": map[string]interface{}{
			"raw":    "This is a test comment",
			"markup": "markdown",
			"html":   "<p>This is a test comment</p>",
		},
		"parent": float64(123456),
	}

	result, err := decodePullRequestsComments(input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(756186348), result.Id)
	assert.Equal(t, "This is a test comment", result.Content.Raw)
	assert.Equal(t, "markdown", result.Content.Markup)
	assert.Equal(t, "<p>This is a test comment</p>", result.Content.Html)
	assert.NotNil(t, result.Parent)
	assert.Equal(t, uint(123456), *result.Parent)
}

func TestDecodePullRequestsComments_ErrorType(t *testing.T) {
	// Response with error type
	input := map[string]interface{}{
		"type": "error",
		"error": map[string]interface{}{
			"message": "Comment not found",
		},
	}

	result, err := decodePullRequestsComments(input)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestDecodePullRequestsComments_NoParent(t *testing.T) {
	// Comment without parent (top-level comment)
	input := map[string]interface{}{
		"id": float64(999),
		"content": map[string]interface{}{
			"raw": "Top-level comment",
		},
	}

	result, err := decodePullRequestsComments(input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(999), result.Id)
	assert.Equal(t, "Top-level comment", result.Content.Raw)
	assert.Nil(t, result.Parent)
}

func TestDecodePullRequestsComments_MinimalData(t *testing.T) {
	// Minimal valid data
	input := map[string]interface{}{
		"id": float64(1),
	}

	result, err := decodePullRequestsComments(input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(1), result.Id)
	assert.Equal(t, "", result.Content.Raw)
}

func TestDecodePullRequestsCommentsList_Success(t *testing.T) {
	// Valid comments list response
	input := map[string]interface{}{
		"page":    float64(1),
		"pagelen": float64(20),
		"size":    float64(3),
		"values": []interface{}{
			map[string]interface{}{
				"id": float64(1),
				"content": map[string]interface{}{
					"raw": "First comment",
				},
			},
			map[string]interface{}{
				"id": float64(2),
				"content": map[string]interface{}{
					"raw": "Second comment",
				},
				"parent": float64(1),
			},
			map[string]interface{}{
				"id": float64(3),
				"content": map[string]interface{}{
					"raw": "Third comment",
				},
			},
		},
	}

	result, err := decodePullRequestsCommentsList(input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(1), result.Page)
	assert.Equal(t, uint(20), result.Pagelen)
	assert.Equal(t, uint(3), result.Size)
	assert.Len(t, result.Items, 3)
	assert.Equal(t, uint(1), result.Items[0].Id)
	assert.Equal(t, "First comment", result.Items[0].Content.Raw)
	assert.Nil(t, result.Items[0].Parent)
	assert.Equal(t, uint(2), result.Items[1].Id)
	assert.Equal(t, "Second comment", result.Items[1].Content.Raw)
	assert.NotNil(t, result.Items[1].Parent)
	assert.Equal(t, uint(1), *result.Items[1].Parent)
}

func TestDecodePullRequestsCommentsList_InvalidFormat(t *testing.T) {
	// Invalid format - not a map
	input := []interface{}{"invalid", "array"}

	result, err := decodePullRequestsCommentsList(input)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Not a valid format")
}

func TestDecodePullRequestsCommentsList_EmptyValues(t *testing.T) {
	// Empty values array
	input := map[string]interface{}{
		"page":    float64(1),
		"pagelen": float64(20),
		"size":    float64(0),
		"values":  []interface{}{},
	}

	result, err := decodePullRequestsCommentsList(input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(1), result.Page)
	assert.Equal(t, uint(20), result.Pagelen)
	assert.Equal(t, uint(0), result.Size)
	assert.Len(t, result.Items, 0)
}

func TestDecodePullRequestsCommentsList_MissingPaginationFields(t *testing.T) {
	// Missing pagination fields - should default to 0
	input := map[string]interface{}{
		"values": []interface{}{
			map[string]interface{}{
				"id": float64(1),
				"content": map[string]interface{}{
					"raw": "Comment",
				},
			},
		},
	}

	result, err := decodePullRequestsCommentsList(input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(0), result.Page)
	assert.Equal(t, uint(0), result.Pagelen)
	assert.Equal(t, uint(0), result.Size)
	assert.Len(t, result.Items, 1)
}

func TestDecodePullRequestsCommentsList_WithErrorInItem(t *testing.T) {
	// One valid comment, one with error type (should be skipped)
	input := map[string]interface{}{
		"page":    float64(1),
		"pagelen": float64(20),
		"size":    float64(2),
		"values": []interface{}{
			map[string]interface{}{
				"id": float64(1),
				"content": map[string]interface{}{
					"raw": "Valid comment",
				},
			},
			map[string]interface{}{
				"type": "error",
				"error": map[string]interface{}{
					"message": "Error in comment",
				},
			},
		},
	}

	result, err := decodePullRequestsCommentsList(input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	// Only the valid comment should be in the list
	assert.Len(t, result.Items, 1)
	assert.Equal(t, uint(1), result.Items[0].Id)
}

func TestDecodePullRequestsCommentsList_InvalidPaginationTypes(t *testing.T) {
	// Invalid pagination field types - should default to 0
	input := map[string]interface{}{
		"page":    "invalid",
		"pagelen": true,
		"size":    []interface{}{},
		"values": []interface{}{
			map[string]interface{}{
				"id": float64(1),
			},
		},
	}

	result, err := decodePullRequestsCommentsList(input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(0), result.Page)
	assert.Equal(t, uint(0), result.Pagelen)
	assert.Equal(t, uint(0), result.Size)
	assert.Len(t, result.Items, 1)
}

func TestDecodePullRequestsList_InvalidPaginationTypes(t *testing.T) {
	// Invalid pagination field types - should default to 0
	input := map[string]interface{}{
		"page":    "invalid",
		"pagelen": nil,
		"size":    map[string]interface{}{},
		"values": []interface{}{
			map[string]interface{}{
				"id": float64(1),
			},
		},
	}

	result, err := decodePullRequestsList(input)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(0), result.Page)
	assert.Equal(t, uint(0), result.Pagelen)
	assert.Equal(t, uint(0), result.Size)
	assert.Len(t, result.Items, 1)
}

func TestDecodePullRequests_ComplexStructure(t *testing.T) {
	// Complex nested structure
	createdTime := time.Now().Add(-24 * time.Hour)
	updatedTime := time.Now()

	input := map[string]interface{}{
		"id":                  float64(456),
		"title":               "Complex PR",
		"description":         "Detailed description",
		"state":               "MERGED",
		"draft":               false,
		"close_source_branch": true,
		"created_on":          createdTime,
		"updated_on":          updatedTime,
		"author": map[string]interface{}{
			"display_name": "Jane Developer",
			"uuid":         "{98765432-9876-9876-9876-987654321098}",
			"type":         "user",
		},
		"reviewers": []interface{}{
			map[string]interface{}{
				"display_name": "Reviewer One",
				"uuid":         "{11111111-1111-1111-1111-111111111111}",
			},
			map[string]interface{}{
				"display_name": "Reviewer Two",
				"uuid":         "{22222222-2222-2222-2222-222222222222}",
			},
		},
	}

	result, err := decodePullRequests(input)
	if err != nil {
		t.Error(err)
	}

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, uint(456), result.ID)
	assert.Equal(t, "Complex PR", result.Title)
	assert.Equal(t, "Detailed description", result.Description)
	assert.Equal(t, "MERGED", result.State)
	assert.False(t, result.Draft)
	assert.True(t, result.CloseSourceBranch)
	assert.NotNil(t, result.CreatedOnTime)
	assert.NotNil(t, result.UpdatedOnTime)
}
