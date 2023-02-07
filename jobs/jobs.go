package jobs

import (
	"strings"

	"github.com/google/go-github/v43/github"
	"github.com/goruha/permbits"
	"github.com/jasonlvhit/gocron"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/terjekv/github-authorized-keys/api"
	"github.com/terjekv/github-authorized-keys/config"
	model "github.com/terjekv/github-authorized-keys/model/linux"
	"github.com/valyala/fasttemplate"
)

const wrapperScriptTpl = `#!/bin/bash
curl http://localhost:{port}/user/$1/authorized_keys
`

func init() {
	viper.SetDefault("ssh_restart_tpl", "/usr/sbin/service ssh force-reload")
	viper.SetDefault("authorized_keys_command_tpl", "/usr/bin/github-authorized-keys")
}

// Run - start scheduled jobs
func Run(cfg config.Config) {
	log.Info("Run syncUsers job on start")
	syncUsers(cfg)

	if cfg.IntegrateWithSSH {
		log.Info("Run ssh integration job on start")
		sshIntegrate(cfg)
	}

	if cfg.Interval != 0 {
		gocron.Every(cfg.Interval).Seconds().Do(syncUsers, cfg)

		// function Start start all the pending jobs
		gocron.Start()
		log.Info("Start jobs scheduler")
	}
}

func syncUsers(cfg config.Config) {
	logger := log.WithFields(log.Fields{"subsystem": "jobs", "job": "syncUsers"})

	c := api.NewGithubClient(cfg.GithubAPIToken, cfg.GithubOrganization)

	if cfg.GithubAdminTeamName != "" {
		team, err := c.GetTeam(cfg.GithubAdminTeamName, cfg.GithubAdminTeamID)
		if err != nil {
			logger.Error(err)
			return
		}

		syncTeamUsers(cfg, c, team, cfg.UserAdminGroups)
	}

	if cfg.GithubUserTeamName != "" {
		team, err := c.GetTeam(cfg.GithubUserTeamName, cfg.GithubUserTeamID)
		if err != nil {
			logger.Error(err)
			return
		}

		syncTeamUsers(cfg, c, team, cfg.UserUserGroups)
	}
}

func syncTeamUsers(cfg config.Config, c *api.GithubClient, team *github.Team, groups []string) {
	logger := log.WithFields(log.Fields{"subsystem": "jobs", "job": "syncTeamUsers"})
	linux := api.NewLinux(cfg.Root)

	log.Info("SyncTeamUsers")

	// Get all GitHub team members
	githubUsers, err := c.GetTeamMembers(team)
	if err != nil {
		logger.Error(err)
		return
	}

	// Track users that were unable to be added to the system
	notCreatedUsers := make([]string, 0)

	for _, githubUser := range githubUsers {
		log.Info(*githubUser.Login)
		linuxUser := model.NewUser(*githubUser.Login, "999", groups, cfg.UserShell)
		// Only add new users
		if !linux.UserExists(linuxUser.Name()) {
			// Create user and track if we failed to create their account
			if err := linux.UserCreate(linuxUser); err != nil {
				logger.Error(err)
				notCreatedUsers = append(notCreatedUsers, linuxUser.Name())
			}
		} else {
			logger.Debugf("User %v exists - skip creation", *githubUser.Login)
		}
	}

}

func sshIntegrate(cfg config.Config) {
	logger := log.WithFields(log.Fields{"subsystem": "jobs", "job": "sshIntegrate"})
	linux := api.NewLinux(cfg.Root)

	// Split listen string by : and get the port
	port := strings.Split(cfg.Listen, ":")[1]

	wrapperScript := fasttemplate.New(wrapperScriptTpl, "{", "}").
		ExecuteString(map[string]interface{}{"port": port})

	cmdFile := viper.GetString("authorized_keys_command_tpl")

	logger.Infof("Ensure file %v", cmdFile)
	linux.FileEnsure(cmdFile, wrapperScript)

	// Should be executable
	logger.Infof("Ensure exec mode for file %v", cmdFile)
	linux.FileModeSet(cmdFile, permbits.PermissionBits(0755))

	logger.Info("Ensure AuthorizedKeysCommand line in sshd_config")
	linux.FileEnsureLineMatch("/etc/ssh/sshd_config", "(?m:^AuthorizedKeysCommand\\s.*$)", "AuthorizedKeysCommand "+cmdFile)

	logger.Info("Ensure AuthorizedKeysCommandUser line in sshd_config")
	linux.FileEnsureLineMatch("/etc/ssh/sshd_config", "(?m:^AuthorizedKeysCommandUser\\s.*$)", "AuthorizedKeysCommandUser nobody")

	logger.Info("Restart ssh")
	output, err := linux.TemplateCommand(viper.GetString("ssh_restart_tpl"), map[string]interface{}{}).CombinedOutput()
	logger.Infof("Output: %v", string(output))
	if err != nil {
		logger.Errorf("Error: %v", err.Error())
	}
}
