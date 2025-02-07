/*
 * Github Authorized Keys - Use GitHub teams to manage system user accounts and authorized_keys
 *
 * Copyright 2016 Cloud Posse, LLC <hello@cloudposse.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package api

import (
	"fmt"
	"os/user"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	model "github.com/terjekv/github-authorized-keys/model/linux"
)

var _ = Describe("Linux", func() {
	Describe("userLookup()", func() {
		Context("call with non-existing user", func() {
			It("should return nil user and error", func() {
				linux := NewLinux("/")

				userName := "testdsadasfsa"

				user, err := linux.userLookup(userName)

				Expect(err).NotTo(BeNil())

				Expect(err.Error()).To(Equal(fmt.Sprintf("user: unknown user %v", userName)))

				Expect(user).To(BeNil())
			})
		})

		Context("call with existing user", func() {
			It("should return valid user", func() {
				linux := NewLinux("/")
				user, err := linux.userLookup("root")

				Expect(err).To(BeNil())

				Expect(user).NotTo(BeNil())

				Expect(user.Gid).To(Equal("0"))
				Expect(user.HomeDir).To(Equal("/root"))
				Expect(user.Name).To(Equal("root"))
				Expect(user.Uid).To(Equal("0"))
				Expect(user.Username).To(Equal("root"))
			})
		})
	})

	Describe("userExists()", func() {
		Context("call with non-existing user", func() {
			It("should return false", func() {
				linux := NewLinux("/")
				isFound := linux.UserExists("testdsadasfsa")
				Expect(isFound).To(BeFalse())
			})
		})

		Context("call with existing user", func() {
			It("should return true", func() {
				linux := NewLinux("/")
				isFound := linux.UserExists("root")
				Expect(isFound).To(BeTrue())
			})
		})
	})

	Describe("userCreate()", func() {
		Context("call without GID", func() {
			var (
				userName model.User
				linux    Linux
			)

			BeforeEach(func() {
				userName = model.NewUser("test", "", []string{"operator", "root"}, "/bin/bash")
				linux = NewLinux("/")
			})

			AfterEach(func() {
				linux.userDelete(userName)
			})

			It("should create valid user", func() {
				err := linux.UserCreate(userName)

				Expect(err).To(BeNil())

				osUser, _ := user.Lookup(userName.Name())

				Expect(osUser.Username).To(Equal(userName.Name()))

				value, _ := strconv.ParseInt(osUser.Gid, 10, 64)
				Expect(value > 0).To(BeTrue())

				gids, _ := osUser.GroupIds()

				for _, group := range userName.Groups() {
					linuxGroup, err := user.LookupGroup(group)
					Expect(err).To(BeNil())
					Expect(gids).To(ContainElement(string(linuxGroup.Gid)))
				}

				shell := linux.userShell(userName.Name())

				Expect(shell).To(Equal(userName.Shell()))
			})
		})

		Context("call with GID", func() {
			var (
				userName model.User
				linux    Linux
			)

			BeforeEach(func() {
				userName = model.NewUser("test", "42", []string{"root"}, "/bin/bash")
				linux = NewLinux("/")
			})

			AfterEach(func() {
				linux.userDelete(userName)
			})

			It("should create valid user", func() {
				err := linux.UserCreate(userName)

				Expect(err).To(BeNil())

				osUser, _ := user.Lookup(userName.Name())

				Expect(osUser.Username).To(Equal(userName.Name()))

				Expect(string(osUser.Gid)).To(Equal(userName.Gid()))

				gids, _ := osUser.GroupIds()

				for _, group := range userName.Groups() {
					linuxGroup, err := user.LookupGroup(group)
					Expect(err).To(BeNil())
					Expect(gids).To(ContainElement(string(linuxGroup.Gid)))
				}

				shell := linux.userShell(userName.Name())

				Expect(shell).To(Equal(userName.Shell()))
			})
		})
	})

	Describe("groupLookup()", func() {
		Context("call with non-existing group", func() {
			It("should return nil group and error", func() {
				linux := NewLinux("/")

				groupName := "testdsadasfsa"

				group, err := linux.groupLookup(groupName)

				Expect(err).NotTo(BeNil())

				Expect(err.Error()).To(Equal(fmt.Sprintf("group: unknown group %v", groupName)))

				Expect(group).To(BeNil())
			})
		})

		Context("call with existing group", func() {
			It("should return valid group", func() {
				linux := NewLinux("/")
				group, err := linux.groupLookup("operator")

				Expect(err).To(BeNil())

				Expect(group).NotTo(BeNil())

				Expect(group.Gid).To(Equal("37"))
				Expect(group.Name).To(Equal("operator"))
			})
		})

		Context("call with existing group with users", func() {
			It("should return valid group", func() {
				linux := NewLinux("/")
				group, err := linux.groupLookup("root")

				Expect(err).To(BeNil())

				Expect(group).NotTo(BeNil())

				Expect(group.Gid).To(Equal("0"))
				Expect(group.Name).To(Equal("root"))
			})
		})
	})

	Describe("groupLookupById()", func() {
		Context("call with non-existing group", func() {
			It("should return nil group and error", func() {
				linux := NewLinux("/")

				groupID := "843"

				group, err := linux.groupLookupByID(groupID)

				Expect(err).NotTo(BeNil())

				Expect(err.Error()).To(Equal(fmt.Sprintf("group: unknown groupid %v", groupID)))

				Expect(group).To(BeNil())
			})
		})

		Context("call with existing group", func() {
			It("should return valid group", func() {
				linux := NewLinux("/")
				group, err := linux.groupLookup("37")

				Expect(err).To(BeNil())

				Expect(group).NotTo(BeNil())

				Expect(group.Gid).To(Equal("37"))
				Expect(group.Name).To(Equal("operator"))
			})
		})

		Context("call with existing group with users", func() {
			It("should return valid group", func() {
				linux := NewLinux("/")
				group, err := linux.groupLookup("0")

				Expect(err).To(BeNil())

				Expect(group).NotTo(BeNil())

				Expect(group.Gid).To(Equal("0"))
				Expect(group.Name).To(Equal("root"))
			})
		})
	})

	Describe("groupExists()", func() {
		Context("call with no existing group", func() {
			It("should return false", func() {
				linux := NewLinux("/")
				isFound := linux.GroupExists("testdsadasfsa")
				Expect(isFound).To(BeFalse())
			})
		})

		Context("call with existing group", func() {
			It("should return true", func() {
				linux := NewLinux("/")
				isFound := linux.GroupExists("operator")
				Expect(isFound).To(BeTrue())
			})
		})
	})

	Describe("groupExistsByID()", func() {
		Context("call with no existing group", func() {
			It("should return false", func() {
				linux := NewLinux("/")
				isFound := linux.groupExistsByID("843")
				Expect(isFound).To(BeFalse())
			})
		})

		Context("call with existing group", func() {
			It("should return true", func() {
				linux := NewLinux("/")
				isFound := linux.groupExistsByID("42")
				Expect(isFound).To(BeTrue())
			})
		})
	})

	Describe("userShell()", func() {
		Context("call with existing user", func() {
			It("should return /bin/bash", func() {
				linux := NewLinux("/")
				shell := linux.userShell("root")
				Expect(shell).To(Equal("/bin/bash"))
			})
		})
	})

})
