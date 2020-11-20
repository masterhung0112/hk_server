package model

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostToJson(t *testing.T) {
	o := Post{Id: NewId(), Message: NewId()}
	j := o.ToJson()
	ro := PostFromJson(strings.NewReader(j))

	assert.NotNil(t, ro)
	assert.Equal(t, &o, ro.Clone())
}

func TestPostFromJsonError(t *testing.T) {
	ro := PostFromJson(strings.NewReader(""))
	assert.Nil(t, ro)
}

func TestPostIsValid(t *testing.T) {
	o := Post{}
	maxPostSize := 10000

	err := o.IsValid(maxPostSize)
	require.NotNil(t, err)

	o.Id = NewId()
	err = o.IsValid(maxPostSize)
	require.NotNil(t, err)

	o.CreateAt = GetMillis()
	err = o.IsValid(maxPostSize)
	require.NotNil(t, err)

	o.UpdateAt = GetMillis()
	err = o.IsValid(maxPostSize)
	require.NotNil(t, err)

	o.UserId = NewId()
	err = o.IsValid(maxPostSize)
	require.NotNil(t, err)

	o.ChannelId = NewId()
	o.RootId = "123"
	err = o.IsValid(maxPostSize)
	require.NotNil(t, err)

	o.RootId = ""
	o.ParentId = "123"
	err = o.IsValid(maxPostSize)
	require.NotNil(t, err)

	o.ParentId = NewId()
	o.RootId = ""
	err = o.IsValid(maxPostSize)
	require.NotNil(t, err)

	o.ParentId = ""
	o.Message = strings.Repeat("0", maxPostSize+1)
	err = o.IsValid(maxPostSize)
	require.NotNil(t, err)

	o.Message = strings.Repeat("0", maxPostSize)
	err = o.IsValid(maxPostSize)
	require.Nil(t, err)

	o.Message = "test"
	err = o.IsValid(maxPostSize)
	require.Nil(t, err)
	o.Type = "junk"
	err = o.IsValid(maxPostSize)
	require.NotNil(t, err)

	o.Type = POST_CUSTOM_TYPE_PREFIX + "type"
	err = o.IsValid(maxPostSize)
	require.Nil(t, err)
}

func TestPostPreSave(t *testing.T) {
	o := Post{Message: "test"}
	o.PreSave()

	require.NotEqual(t, 0, o.CreateAt)

	past := GetMillis() - 1
	o = Post{Message: "test", CreateAt: past}
	o.PreSave()

	require.LessOrEqual(t, o.CreateAt, past)

	o.Etag()
}

func TestPostIsSystemMessage(t *testing.T) {
	post1 := Post{Message: "test_1"}
	post1.PreSave()

	require.False(t, post1.IsSystemMessage())

	post2 := Post{Message: "test_2", Type: POST_JOIN_LEAVE}
	post2.PreSave()

	require.True(t, post2.IsSystemMessage())
}
