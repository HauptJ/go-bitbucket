package tests

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ktrysmt/go-bitbucket"
)

func TestProfile(t *testing.T) {

	user := getUsername()
	pass := getPassword()

	c, err := bitbucket.NewBasicAuth(user, pass)
	if err != nil {
		t.Error(err)
	}

	res, err := c.User.Profile()

	assert.NoError(t, err)
	assert.NotNil(t, res)

	resStr := res.String()
	assert.Contains(t, resStr, "User: "+res.DisplayName)
}

func TestString(t *testing.T) {
	u := &bitbucket.User{DisplayName: "John Doe", AccountId: "12345"}
	assert.Equal(t, "User: John Doe, Account ID: 12345", u.String())
}

func getUsername() string {
	ev := os.Getenv("BITBUCKET_TEST_USERNAME")
	if ev != "" {
		return ev
	}

	return "example-username"
}

func getPassword() string {
	ev := os.Getenv("BITBUCKET_TEST_PASSWORD")
	if ev != "" {
		return ev
	}

	return "password"
}
