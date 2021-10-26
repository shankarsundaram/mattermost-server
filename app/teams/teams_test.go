// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package teams

import (
	"strings"
	"testing"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/stretchr/testify/require"
)

func TestCreateTeam(t *testing.T) {
	th := Setup(t)
	defer th.TearDown()

	id := model.NewId()
	team := &model.Team{
		DisplayName: "dn_" + id,
		Name:        "name" + id,
		Email:       "success+" + id + "@simulator.amazonses.com",
		Type:        model.TeamOpen,
	}

	_, err := th.service.CreateTeam(team)
	require.NoError(t, err, "Should create a new team")

	_, err = th.service.CreateTeam(team)
	require.Error(t, err, "Should not create a new team - team already exist")
}

func TestJoinUserToTeam(t *testing.T) {
	th := Setup(t)
	defer th.TearDown()

	id := model.NewId()
	team := &model.Team{
		DisplayName: "dn_" + id,
		Name:        "name" + id,
		Email:       "success+" + id + "@simulator.amazonses.com",
		Type:        model.TeamOpen,
	}

	_, err := th.service.CreateTeam(team)
	require.NoError(t, err, "Should create a new team")

	maxUsersPerTeam := th.service.config().TeamSettings.MaxUsersPerTeam
	defer func() {
		th.UpdateConfig(func(cfg *model.Config) { cfg.TeamSettings.MaxUsersPerTeam = maxUsersPerTeam })
		th.DeleteTeam(team)
	}()
	one := 1
	th.UpdateConfig(func(cfg *model.Config) { cfg.TeamSettings.MaxUsersPerTeam = &one })

	t.Run("new join", func(t *testing.T) {
		user := model.User{Email: strings.ToLower(model.NewId()) + "success+test@example.com", Nickname: "Darth Vader", Username: "vader" + model.NewId(), Password: "passwd1", AuthService: ""}
		ruser := th.CreateUser(&user)
		defer th.DeleteUser(&user)

		_, alreadyAdded, err := th.service.JoinUserToTeam(team, ruser)
		require.False(t, alreadyAdded, "Should return already added equal to false")
		require.NoError(t, err)
	})

	t.Run("join when you are a member", func(t *testing.T) {
		user := model.User{Email: strings.ToLower(model.NewId()) + "success+test@example.com", Nickname: "Darth Vader", Username: "vader" + model.NewId(), Password: "passwd1", AuthService: ""}
		ruser := th.CreateUser(&user)
		defer th.DeleteUser(&user)

		_, _, err := th.service.JoinUserToTeam(team, ruser)
		require.NoError(t, err)

		_, alreadyAdded, err := th.service.JoinUserToTeam(team, ruser)
		require.True(t, alreadyAdded, "Should return already added")
		require.NoError(t, err)
	})

	t.Run("re-join after leaving", func(t *testing.T) {
		user := model.User{Email: strings.ToLower(model.NewId()) + "success+test@example.com", Nickname: "Darth Vader", Username: "vader" + model.NewId(), Password: "passwd1", AuthService: ""}
		ruser := th.CreateUser(&user)
		defer th.DeleteUser(&user)

		member, _, err := th.service.JoinUserToTeam(team, ruser)
		require.NoError(t, err)
		err = th.service.RemoveTeamMember(member)
		require.NoError(t, err)

		_, alreadyAdded, err := th.service.JoinUserToTeam(team, ruser)
		require.False(t, alreadyAdded, "Should return already added equal to false")
		require.NoError(t, err)
	})

	t.Run("new join with limit problem", func(t *testing.T) {
		user1 := model.User{Email: strings.ToLower(model.NewId()) + "success+test@example.com", Nickname: "Darth Vader", Username: "vader" + model.NewId(), Password: "passwd1", AuthService: ""}
		ruser1 := th.CreateUser(&user1)
		user2 := model.User{Email: strings.ToLower(model.NewId()) + "success+test@example.com", Nickname: "Darth Vader", Username: "vader" + model.NewId(), Password: "passwd1", AuthService: ""}
		ruser2 := th.CreateUser(&user2)

		defer th.DeleteUser(&user1)
		defer th.DeleteUser(&user2)

		_, _, err := th.service.JoinUserToTeam(team, ruser1)
		require.NoError(t, err)

		_, _, err = th.service.JoinUserToTeam(team, ruser2)
		require.Error(t, err, "Should fail")
	})

	t.Run("re-join alfter leaving with limit problem", func(t *testing.T) {
		user1 := model.User{Email: strings.ToLower(model.NewId()) + "success+test@example.com", Nickname: "Darth Vader", Username: "vader" + model.NewId(), Password: "passwd1", AuthService: ""}
		ruser1 := th.CreateUser(&user1)

		user2 := model.User{Email: strings.ToLower(model.NewId()) + "success+test@example.com", Nickname: "Darth Vader", Username: "vader" + model.NewId(), Password: "passwd1", AuthService: ""}
		ruser2 := th.CreateUser(&user2)

		defer th.DeleteUser(&user1)
		defer th.DeleteUser(&user2)

		member, _, err := th.service.JoinUserToTeam(team, ruser1)
		require.NoError(t, err)
		err = th.service.RemoveTeamMember(member)
		require.NoError(t, err)
		_, _, err = th.service.JoinUserToTeam(team, ruser2)
		require.NoError(t, err)

		_, _, err = th.service.JoinUserToTeam(team, ruser1)
		require.Error(t, err, "Should fail")
	})
}

func TestUpdateTeamMemberRolesChangingGuest(t *testing.T) {
	th := Setup(t)
	defer th.TearDown()

	id := model.NewId()
	team := &model.Team{
		DisplayName: "dn_" + id,
		Name:        "name" + id,
		Email:       "success+" + id + "@simulator.amazonses.com",
		Type:        model.TeamOpen,
	}

	_, err := th.service.CreateTeam(team)
	require.NoError(t, err, "Should create a new team")

	user := model.User{Email: strings.ToLower(model.NewId()) + "success+test@example.com", Nickname: "Darth Vader", Username: "vader" + model.NewId(), Password: "passwd1", AuthService: ""}
	ruser := th.CreateUser(&user)

	_, _, err = th.service.JoinUserToTeam(team, ruser)
	require.NoError(t, err)

	user2 := model.User{Email: strings.ToLower(model.NewId()) + "success+test+2@example.com", Nickname: "Luke", Username: "luke" + model.NewId(), Password: "passwd1", AuthService: ""}
	ruser2 := th.CreateUser(&user2)

	t.Run("not a team member", func(t *testing.T) {
		_, err = th.service.UpdateTeamMemberRoles(team.Id, ruser2.Id, "team_admin")
		require.Error(t, err, "Should fail when try to modify the a non member")
	})

	t.Run("no schemes for team", func(t *testing.T) {
		team.SchemeId = model.NewString("foo")
		_, err = th.service.UpdateTeam(team, UpdateOptions{Imported: true})
		require.NoError(t, err)
		defer func() {
			team.SchemeId = nil
			_, err = th.service.UpdateTeam(team, UpdateOptions{Imported: true})
			require.NoError(t, err)
		}()

		_, err = th.service.UpdateTeamMemberRoles(team.Id, ruser.Id, "team_admin")
		require.Error(t, err, "Should fail if there is no scheme")
	})

	t.Run("role not found", func(t *testing.T) {
		_, err = th.service.UpdateTeamMemberRoles(team.Id, ruser.Id, "new_role")
		require.Error(t, err, "Should fail if there is no such role")
	})

	t.Run("managed role", func(t *testing.T) {
		_, err = th.service.roleStore.Save(&model.Role{
			Name:          "foo_role",
			SchemeManaged: true,
			DisplayName:   "managed role",
			Description:   "desc",
			Permissions:   []string{"manage_system"},
		})
		require.NoError(t, err)

		_, err = th.service.UpdateTeamMemberRoles(team.Id, ruser.Id, "foo_role")
		require.Error(t, err, "Should fail if the role is managed")
	})
}

func TestUpdateTeamMemberSchemeRoles(t *testing.T) {
	th := Setup(t)
	defer th.TearDown()

	id := model.NewId()
	team := &model.Team{
		DisplayName: "dn_" + id,
		Name:        "name" + id,
		Email:       "success+" + id + "@simulator.amazonses.com",
		Type:        model.TeamOpen,
	}

	_, err := th.service.CreateTeam(team)
	require.NoError(t, err, "Should create a new team")

	user := model.User{Email: strings.ToLower(model.NewId()) + "success+test@example.com", Nickname: "Darth Vader", Username: "vader" + model.NewId(), Password: "passwd1", AuthService: ""}
	ruser := th.CreateUser(&user)

	_, _, err = th.service.JoinUserToTeam(team, ruser)
	require.NoError(t, err)

	user2 := model.User{Email: strings.ToLower(model.NewId()) + "success+test+2@example.com", Nickname: "Luke", Username: "luke" + model.NewId(), Password: "passwd1", AuthService: ""}
	ruser2 := th.CreateUser(&user2)

	t.Run("not a team member", func(t *testing.T) {
		_, err := th.service.UpdateTeamMemberSchemeRoles(team.Id, ruser2.Id, true, true, false, false)
		require.Error(t, err, "Should fail when try to modify the a non member")
	})

	t.Run("invalid roles", func(t *testing.T) {
		_, err := th.service.UpdateTeamMemberSchemeRoles(team.Id, ruser.Id, true, true, true, false)
		require.Error(t, err, "Should fail when try to make member both user and guest")
	})

	t.Run("update role to guest", func(t *testing.T) {
		_, err := th.service.UpdateTeamMemberSchemeRoles(team.Id, ruser.Id, true, false, false, false)
		require.Nil(t, err, "Should not fail when try to make member a guest")
	})
}