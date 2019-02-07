package api

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gopkg.in/ldap.v3"
	"testing"
)

type MockClient struct {
	ldap.Client
}

func (c *MockClient) Bind(dn, password string) error {
	if password == "fail" {
		return fmt.Errorf("Failed !")
	}
	return nil
}

func (c *MockClient) Search(r *ldap.SearchRequest) (*ldap.SearchResult, error) {
	if r.BaseDN == "DC=foo,DC=bar" {
		return &ldap.SearchResult{}, nil
	}
	return &ldap.SearchResult{
		Entries: []*ldap.Entry{&ldap.Entry{DN: "DC=foo, DC=baz"}},
	}, nil
}

func TestLDAPAuth(t *testing.T) {
	auth, _ := NewLDAPAuth("uri", "baseDN", "bindDN", "user", "fail")
	client := &MockClient{}
	// test bind failure
	res, err := auth.doAuth("foo", "bar", client)
	assert.Equal(t, res, false)
	assert.NotNil(t, err)
	// test no entries
	auth, _ = NewLDAPAuth("uri", "DC=foo,DC=bar", "bindDN", "user", "pass")
	res, err = auth.doAuth("foo", "bar", client)
	assert.Equal(t, res, false)
	assert.NotNil(t, err)
	// test auth failure
	auth, _ = NewLDAPAuth("uri", "baseDN", "bindDN", "user", "pass")
	res, err = auth.doAuth("foo", "fail", client)
	assert.Equal(t, res, false)
	assert.Nil(t, err)
	// test auth success
	res, err = auth.doAuth("foo", "bar", client)
	assert.Equal(t, res, true)
	assert.Nil(t, err)
}
