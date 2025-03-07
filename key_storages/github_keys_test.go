package keyStorages

import (
	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

var _ = Describe("GithubKeys as backend storage", func() {
	var (
		validToken         string
		validOrg           string
		validAdminTeamName string
		validAdminTeamID   int
		validUserTeamName  string
		validUserTeamID    int
		validUser          string
	)

	BeforeEach(func() {
		validToken = viper.GetString("github_api_token")
		validOrg = viper.GetString("github_organization")
		validAdminTeamName = viper.GetString("github_admin_team")
		validAdminTeamID = viper.GetInt("github_admin_team_id")
		validUserTeamName = viper.GetString("github_user_team")
		validUserTeamID = viper.GetInt("github_admin_user_id")
		validUser = viper.GetString("github_user")
	})

	Describe("when github.com up", func() {
		var c *GithubKeys

		BeforeEach(func() {
			c = NewGithubKeys(validToken, validOrg, validAdminTeamName, validAdminTeamID, validUserTeamName, validUserTeamID)
		})

		Context("backend have valid value", func() {

			It("should return valid value", func() {

				keys, err := c.Get(validUser)

				Expect(err).To(BeNil())
				Expect(keys).NotTo(Equal(""))
			})
		})

		Context("backend does not have valid value", func() {
			It("should return empty value and valid error", func() {

				keys, err := c.Get("djahsjdhadafdsgfhdgahfjd")

				Expect(err).NotTo(BeNil())
				Expect(err).To(Equal(ErrStorageKeyNotFound))
				Expect(keys).To(Equal(""))
			})
		})
	})

	Describe("when github.com down", func() {

		var c *GithubKeys

		BeforeEach(func() {
			httpmock.Activate()
			c = NewGithubKeys(validToken, validOrg, validAdminTeamName, validAdminTeamID, validUserTeamName, validUserTeamID)

		})

		AfterEach(func() {
			defer httpmock.DeactivateAndReset()
		})

		It("should return valid error and empty string", func() {
			keys, err := c.Get(validUser)

			Expect(err).NotTo(BeNil())
			Expect(err).NotTo(Equal(ErrStorageKeyNotFound))
			Expect(keys).To(Equal(""))
		})
	})

})
