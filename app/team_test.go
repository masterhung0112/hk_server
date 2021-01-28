package app

import (
	"testing"

	"github.com/masterhung0112/hk_server/model"
	"github.com/stretchr/testify/require"
)

func TestCreateTeam(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	id := model.NewId()
	team := &model.Team{
		DisplayName: "dn_" + id,
		Name:        "name" + id,
		Email:       "success+" + id + "@simulator.amazonses.com",
		Type:        model.TEAM_OPEN,
	}

	_, err := th.App.CreateTeam(team)
	require.Nil(t, err, "Should create a new team")

	_, err = th.App.CreateTeam(th.BasicTeam)
	require.NotNil(t, err, "Should not create a new team - team already exist")
}
