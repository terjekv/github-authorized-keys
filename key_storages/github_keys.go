package keyStorages

import (
	"errors"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/terjekv/github-authorized-keys/api"
)

// GithubKeys - github api as key storage
type GithubKeys struct {
	client      *api.GithubClient
	Adminteam   string
	AdminteamID int
	Userteam    string
	UserteamID  int
}

// Get - fetch {user} ssh keys
func (s *GithubKeys) Get(user string) (value string, err error) {
	defer func() {
		if r := recover(); r != nil {
			value = ""
			err = ErrStorageConnectionFailed
		}
	}()

	value = ""

	logger := log.WithFields(log.Fields{"class": "github_keys", "method": "Get"})
	log.SetLevel(log.DebugLevel)

	logger.Debugf("starting lookup %v", user)

	logger.Debugf("checking admin membership")
	aisMember, _ := isMemberOf(s, user, s.Adminteam, s.AdminteamID)

	if !aisMember {
		logger.Debugf("checking user membership")
		uisMember, _ := isMemberOf(s, user, s.Adminteam, s.AdminteamID)
		if !uisMember {
			//			uerr = ErrStorageKeyNotFound
			logger.Debugf("no memberships for %v", user)
			return
		}
	}

	// we have some membership, get keys etc.
	keys, err := s.client.GetKeys(user)

	if err == nil {
		result := []string{}
		for _, value := range keys {
			result = append(result, *value.Key)
		}
		value = strings.Join(result, "\n")

	} else if err == api.ErrorGitHubNotFound {
		err = ErrStorageKeyNotFound
	} else {
		err = errors.New("access denied")
	}

	return
}

func isMemberOf(s *GithubKeys, user string, teamname string, teamid int) (isMember bool, err error) {
	logger := log.WithFields(log.Fields{"class": "github_keys", "method": "isMemberOfAny"})
	log.SetLevel(log.DebugLevel)

	logger.Debugf("fetching team %v/%d", teamname, teamid)
	team, err := s.client.GetTeam(teamname, teamid)
	if err != nil {
		if err == api.ErrorGitHubConnectionFailed {
			err = ErrStorageConnectionFailed
		} else {
			err = ErrStorageKeyNotFound
		}
		return false, err
	}

	logger.Debugf("checking is %v is in team %v", user, teamname)
	isMember, mem_err := s.client.IsTeamMember(user, team)
	if mem_err != nil {
		if mem_err == api.ErrorGitHubConnectionFailed {
			mem_err = ErrStorageConnectionFailed
		} else {
			mem_err = ErrStorageKeyNotFound
		}
		logger.Debugf("looks like is %v is a member of team %v!", user, teamname)
	}
	return isMember, nil
}

// NewGithubKeys - constructor for github key storage
func NewGithubKeys(token, owner, Adminteam string, AdminteamID int, Userteam string, UserteamID int) *GithubKeys {
	return &GithubKeys{
		client:    api.NewGithubClient(token, owner),
		Adminteam: Adminteam, AdminteamID: AdminteamID,
		Userteam: Userteam, UserteamID: UserteamID,
	}
}
