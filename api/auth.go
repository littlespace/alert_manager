package api

import (
	"crypto/tls"
	"fmt"
	"github.com/golang/glog"
	"gopkg.in/ldap.v3"
	"net/url"
	"os"
	"strings"
)

const (
	ldapBindUserEnv = "AM_LDAP_BIND_USER"
	ldapBindPassEnv = "AM_LDAP_BIND_PASSWORD"
)

type LDAPAuth struct {
	URL          string
	BindDN       string
	BaseDN       string
	BindUser     string
	BindPassword string
	secure       bool
}

func NewLDAPAuth(uri, baseDN, bindDN, bindUser, bindPass string) (*LDAPAuth, error) {
	if bindUser == "" || bindPass == "" {
		bindUser = os.Getenv(ldapBindUserEnv)
		bindPass = os.Getenv(ldapBindPassEnv)
	}
	if bindUser == "" || bindPass == "" {
		return nil, fmt.Errorf("Unable to get ldap bind credentials")
	}
	l := &LDAPAuth{BindDN: bindDN, BaseDN: baseDN, BindUser: bindUser, BindPassword: bindPass}
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "ldaps" {
		l.secure = true
	}
	l.URL = u.Host
	return l, nil
}

func (l *LDAPAuth) Authenticate(username, password string) (bool, error) {
	// connect to ldap server
	c, err := ldap.Dial("tcp", l.URL)
	if err != nil {
		return false, fmt.Errorf("failed to connect to ldap: %v", err)
	}
	defer c.Close()
	return l.doAuth(username, password, c)
}

func (l *LDAPAuth) doAuth(username, password string, c ldap.Client) (bool, error) {
	if l.secure {
		err := c.StartTLS(&tls.Config{InsecureSkipVerify: true})
		if err != nil {
			return false, err
		}
	}
	bindDN := strings.Join([]string{fmt.Sprintf("CN=%s", l.BindUser), l.BindDN}, ",")
	// try bind to ldap server
	if err := c.Bind(bindDN, l.BindPassword); err != nil {
		return false, fmt.Errorf("Failed to bind: %v", err)
	}
	// Search for the given username
	searchRequest := ldap.NewSearchRequest(
		l.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectCategory=person)(objectClass=user)(sAMAccountName=%s))", username),
		[]string{"dn"},
		nil,
	)
	sr, err := c.Search(searchRequest)
	if err != nil {
		return false, err
	}
	if len(sr.Entries) != 1 {
		return false, fmt.Errorf("User does not exist or too many entries returned")
	}
	userdn := sr.Entries[0].DN
	// Bind as the user to verify their password
	err = c.Bind(userdn, password)
	if err != nil {
		glog.Errorf("User Authentication failed for %s", username)
		return false, nil
	}
	return true, nil
}
