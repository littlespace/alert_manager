package models

import (
	"database/sql"
	"encoding/json"
)

var (
	QueryInsertTeam         = "INSERT INTO teams (name, organization) VALUES (:name, :organization) RETURNING id"
	QueryDeleteTeam         = "DELETE FROM teams WHERE id=$1"
	QueryInsertUser         = "INSERT INTO users (name, team_id) VALUES (:name, :team_id) RETURNING id"
	QueryDeleteUser         = "DELETE FROM users WHERE id=$1"
	QueryDeleteUsersForTeam = "DELETE FROM USERS WHERE team_id=$1"

	QuerySelectTeams = "SELECT * FROM teams"
	QuerySelectUsers = `
		SELECT users.*, teams.id "team.id", teams.name "team.name", teams.organization "team.organization"
		FROM users
		JOIN teams ON users.team_id = teams.id"
	`
)

type Team struct {
	Id           int64
	Name         string
	Organization sql.NullString
}

func (t *Team) MarshalJSON() ([]byte, error) {
	tm := struct {
		Id           int64
		Name         string
		Organization string
	}{Id: t.Id, Name: t.Name, Organization: t.Organization.String}
	return json.Marshal(&tm)
}

type User struct {
	Id     int64
	Name   string
	TeamId int64 `db:"team_id"`
	Team   `db:"team"`
}

func NewUser(name string, teamId int64) *User {
	return &User{Name: name, TeamId: teamId}
}

type Teams []*Team
type Users []*User

func (t Teams) Contains(teamName string) bool {
	for _, team := range t {
		if team.Name == teamName {
			return true
		}
	}
	return false
}

func (tx *Tx) SelectTeams(query string, args ...interface{}) (Teams, error) {
	var teams Teams
	err := tx.Select(&teams, query, args...)
	return teams, err
}

func (tx *Tx) SelectUsers(query string, args ...interface{}) (Users, error) {
	var users []*User
	err := tx.Select(&users, query, args...)
	return users, err
}
