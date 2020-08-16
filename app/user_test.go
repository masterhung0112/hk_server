package app

import (
	"github.com/stretchr/testify/require"
	"strings"
	"github.com/masterhung0112/go_server/model"
	"testing"
)

func TestCreateUser(t *testing.T) {
  th := Setup(t)
  defer th.TearDown()

  user := model.User{
    Email: strings.ToLower(model.NewId()) + "success+test@example.com",
    Username: "vader" + model.NewId(),
    Password: "Passwds1@",
  }

  t.Run("Valid user", func (t *testing.T) {
    user := user
    ruser, err := th.App.CreateUser(&user)
    require.Nil(t, err, "Should success to create user")
    require.NotNil(t, ruser, "Should return the new ruser")
  })

  t.Run("Empty Username success", func (t *testing.T) {
    user := user
    user.Username = ""
    user.Email = strings.ToLower(model.NewId()) + "success+test@example.com"
    _, err := th.App.CreateUser(&user)
		require.Nil(t, err, "Should success to create user without Username value")
  })
}