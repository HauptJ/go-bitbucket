package tests

import (
	"fmt"
	"strconv"
	"testing"

	go_bitbucket "github.com/ktrysmt/go-bitbucket"
	"github.com/stretchr/testify/assert"
)

func TestListPullRequestsListObjs(t *testing.T) {

	client, err := setupBasicAuthTest(t)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotNil(t, client, "Client should not be nil after setup")

	inPullRequestOpts := &go_bitbucket.PullRequestsOptions{
		Owner:    ownerEnv,
		RepoSlug: repoEnv,
		States:   []string{"DECLINED"},
	}

	actualPullRequests, err := client.Repositories.PullRequests.ListObjs(inPullRequestOpts)
	assert.NoError(t, err)
	assert.NotNil(t, actualPullRequests, "Pull requests should not be nil")
	assert.Len(t, actualPullRequests.Items, 1, "Expected one pull request")
	assert.NotEmpty(t, actualPullRequests.Items[0].UpdatedOnTime, "Pull request updatedOnTime not be empty")
	assert.Contains(t, actualPullRequests.Items[0].UpdatedOnTime.String(), "2026", "Pull request updatedOnTime should contain 2026")
	//assert.Len(t, actualPullRequests.Items[0].Comments, 1, "Expected comments on pull request")
}

func TestListPullRequestsList(t *testing.T) {

	client, err := setupBasicAuthTest(t)

	inPullRequestOpts := &go_bitbucket.PullRequestsOptions{
		Owner:    ownerEnv,
		RepoSlug: repoEnv,
		States:   []string{"DECLINED"},
	}

	if err != nil {
		t.Error(err)
	}
	assert.NotNil(t, client, "Client should not be nil after setup")

	actualPullRequests, err := client.Repositories.PullRequests.List(inPullRequestOpts)
	assert.NoError(t, err)
	assert.NotNil(t, actualPullRequests, "Pull requests should not be nil")

	for key, value := range actualPullRequests.(map[string]interface{}) {
		fmt.Println(key, value)
		if key == "values" {
			for key, value := range value.([]interface{}) {
				fmt.Println(key, value)
				assert.Equal(t, float64(1), value.(map[string]interface{})["id"])
				assert.Equal(t, "getauthtoken - test pull request", value.(map[string]interface{})["title"])
				assert.Equal(t, "test pull request", value.(map[string]interface{})["description"])
				assert.Equal(t, "2026-02-09T17:23:41.159197+00:00", value.(map[string]interface{})["created_on"].(string))
				assert.Greater(t, value.(map[string]interface{})["updated_on"], value.(map[string]interface{})["created_on"])
				assert.Equal(t, false, value.(map[string]interface{})["draft"])
				assert.Equal(t, "DECLINED", value.(map[string]interface{})["state"])
				assert.Equal(t, "Joshua Haupt", value.(map[string]interface{})["author"].(map[string]interface{})["display_name"])
			}
		}
	}
}

func TestGetPullRequestCommentsObjs(t *testing.T) {
	client, err := setupBasicAuthTest(t)
	if err != nil {
		t.Error(err)
	}
	assert.NotNil(t, client, "Client should not be nil after setup")

	inPullRequestsCommentOpts := &go_bitbucket.PullRequestCommentOptions{
		Owner:         ownerEnv,
		RepoSlug:      repoEnv,
		PullRequestID: "1",
		CommentId:     "756186348",
	}

	actualPullRequestsComments, err := client.Repositories.PullRequests.GetCommentObj(inPullRequestsCommentOpts)
	if err != nil {
		t.Error(err)
	}
	assert.NotNil(t, actualPullRequestsComments, "Pull requests comments should not be nil")
	assert.Contains(t, actualPullRequestsComments.Content.Raw, "test", "Expected comment on pull request")
}

func TestGetPullRequestComments_Add_Delete(t *testing.T) {
	client, err := setupBasicAuthTest(t)
	if err != nil {
		t.Error(err)
	}
	assert.NotNil(t, client, "Client should not be nil after setup")

	inPullRequestsCommentOpts := &go_bitbucket.PullRequestCommentOptions{
		Owner:         ownerEnv,
		RepoSlug:      repoEnv,
		PullRequestID: "1",
		Content:       "test",
	}

	newPullRequestsComment, err := client.Repositories.PullRequests.AddCommentObj(*inPullRequestsCommentOpts)
	if err != nil {
		t.Fatal(err)
	}

	deleteCommentOpts := &go_bitbucket.PullRequestCommentOptions{
		Owner:         ownerEnv,
		RepoSlug:      repoEnv,
		CommentId:     strconv.Itoa(int(newPullRequestsComment.Id)),
		PullRequestID: "1",
	}

	_, err = client.Repositories.PullRequests.DeleteComment(deleteCommentOpts)
	if err != nil {
		t.Error(err)
	}
}
