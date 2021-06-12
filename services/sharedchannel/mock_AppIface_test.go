// Code generated by mockery v1.0.0. DO NOT EDIT.

// Regenerate this file using `make sharedchannel-mocks`.

package sharedchannel

import (
	filestore "github.com/masterhung0112/hk_server/v5/shared/filestore"
	mock "github.com/stretchr/testify/mock"

	model "github.com/masterhung0112/hk_server/v5/model"
)

// MockAppIface is an autogenerated mock type for the AppIface type
type MockAppIface struct {
	mock.Mock
}

// AddUserToChannel provides a mock function with given fields: user, channel, skipTeamMemberIntegrityCheck
func (_m *MockAppIface) AddUserToChannel(user *model.User, channel *model.Channel, skipTeamMemberIntegrityCheck bool) (*model.ChannelMember, *model.AppError) {
	ret := _m.Called(user, channel, skipTeamMemberIntegrityCheck)

	var r0 *model.ChannelMember
	if rf, ok := ret.Get(0).(func(*model.User, *model.Channel, bool) *model.ChannelMember); ok {
		r0 = rf(user, channel, skipTeamMemberIntegrityCheck)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.ChannelMember)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(*model.User, *model.Channel, bool) *model.AppError); ok {
		r1 = rf(user, channel, skipTeamMemberIntegrityCheck)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// AddUserToTeamByTeamId provides a mock function with given fields: c, teamId, user
func (_m *MockAppIface) AddUserToTeamByTeamId(c *request.Context, teamId string, user *model.User) *model.AppError {
	ret := _m.Called(c, teamId, user)

	var r0 *model.AppError
	if rf, ok := ret.Get(0).(func(*request.Context, string, *model.User) *model.AppError); ok {
		r0 = rf(c, teamId, user)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AppError)
		}
	}

	return r0
}

// CreateChannelWithUser provides a mock function with given fields: c, channel, userId
func (_m *MockAppIface) CreateChannelWithUser(c *request.Context, channel *model.Channel, userId string) (*model.Channel, *model.AppError) {
	ret := _m.Called(c, channel, userId)

	var r0 *model.Channel
	if rf, ok := ret.Get(0).(func(*request.Context, *model.Channel, string) *model.Channel); ok {
		r0 = rf(c, channel, userId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Channel)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(*request.Context, *model.Channel, string) *model.AppError); ok {
		r1 = rf(c, channel, userId)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// CreatePost provides a mock function with given fields: c, post, channel, triggerWebhooks, setOnline
func (_m *MockAppIface) CreatePost(c *request.Context, post *model.Post, channel *model.Channel, triggerWebhooks bool, setOnline bool) (*model.Post, *model.AppError) {
	ret := _m.Called(c, post, channel, triggerWebhooks, setOnline)

	var r0 *model.Post
	if rf, ok := ret.Get(0).(func(*request.Context, *model.Post, *model.Channel, bool, bool) *model.Post); ok {
		r0 = rf(c, post, channel, triggerWebhooks, setOnline)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Post)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(*request.Context, *model.Post, *model.Channel, bool, bool) *model.AppError); ok {
		r1 = rf(c, post, channel, triggerWebhooks, setOnline)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// CreateUploadSession provides a mock function with given fields: us
func (_m *MockAppIface) CreateUploadSession(us *model.UploadSession) (*model.UploadSession, *model.AppError) {
	ret := _m.Called(us)

	var r0 *model.UploadSession
	if rf, ok := ret.Get(0).(func(*model.UploadSession) *model.UploadSession); ok {
		r0 = rf(us)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.UploadSession)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(*model.UploadSession) *model.AppError); ok {
		r1 = rf(us)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// DeletePost provides a mock function with given fields: postID, deleteByID
func (_m *MockAppIface) DeletePost(postID string, deleteByID string) (*model.Post, *model.AppError) {
	ret := _m.Called(postID, deleteByID)

	var r0 *model.Post
	if rf, ok := ret.Get(0).(func(string, string) *model.Post); ok {
		r0 = rf(postID, deleteByID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Post)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string, string) *model.AppError); ok {
		r1 = rf(postID, deleteByID)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// DeleteReactionForPost provides a mock function with given fields: c, reaction
func (_m *MockAppIface) DeleteReactionForPost(c *request.Context, reaction *model.Reaction) *model.AppError {
	ret := _m.Called(c, reaction)

	var r0 *model.AppError
	if rf, ok := ret.Get(0).(func(*request.Context, *model.Reaction) *model.AppError); ok {
		r0 = rf(c, reaction)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AppError)
		}
	}

	return r0
}

// FileReader provides a mock function with given fields: path
func (_m *MockAppIface) FileReader(path string) (filestore.ReadCloseSeeker, *model.AppError) {
	ret := _m.Called(path)

	var r0 filestore.ReadCloseSeeker
	if rf, ok := ret.Get(0).(func(string) filestore.ReadCloseSeeker); ok {
		r0 = rf(path)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(filestore.ReadCloseSeeker)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(string) *model.AppError); ok {
		r1 = rf(path)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetOrCreateDirectChannel provides a mock function with given fields: c, userId, otherUserId, channelOptions
func (_m *MockAppIface) GetOrCreateDirectChannel(c *request.Context, userId string, otherUserId string, channelOptions ...model.ChannelOption) (*model.Channel, *model.AppError) {
	_va := make([]interface{}, len(channelOptions))
	for _i := range channelOptions {
		_va[_i] = channelOptions[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, c, userId, otherUserId)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *model.Channel
	if rf, ok := ret.Get(0).(func(*request.Context, string, string, ...model.ChannelOption) *model.Channel); ok {
		r0 = rf(c, userId, otherUserId, channelOptions...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Channel)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(*request.Context, string, string, ...model.ChannelOption) *model.AppError); ok {
		r1 = rf(c, userId, otherUserId, channelOptions...)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// GetProfileImage provides a mock function with given fields: user
func (_m *MockAppIface) GetProfileImage(user *model.User) ([]byte, bool, *model.AppError) {
	ret := _m.Called(user)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(*model.User) []byte); ok {
		r0 = rf(user)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 bool
	if rf, ok := ret.Get(1).(func(*model.User) bool); ok {
		r1 = rf(user)
	} else {
		r1 = ret.Get(1).(bool)
	}

	var r2 *model.AppError
	if rf, ok := ret.Get(2).(func(*model.User) *model.AppError); ok {
		r2 = rf(user)
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).(*model.AppError)
		}
	}

	return r0, r1, r2
}

// InvalidateCacheForUser provides a mock function with given fields: userID
func (_m *MockAppIface) InvalidateCacheForUser(userID string) {
	_m.Called(userID)
}

// MentionsToTeamMembers provides a mock function with given fields: message, teamID
func (_m *MockAppIface) MentionsToTeamMembers(message string, teamID string) model.UserMentionMap {
	ret := _m.Called(message, teamID)

	var r0 model.UserMentionMap
	if rf, ok := ret.Get(0).(func(string, string) model.UserMentionMap); ok {
		r0 = rf(message, teamID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(model.UserMentionMap)
		}
	}

	return r0
}

// NotifySharedChannelUserUpdate provides a mock function with given fields: user
func (_m *MockAppIface) NotifySharedChannelUserUpdate(user *model.User) {
	_m.Called(user)
}

// PatchChannelModerationsForChannel provides a mock function with given fields: channel, channelModerationsPatch
func (_m *MockAppIface) PatchChannelModerationsForChannel(channel *model.Channel, channelModerationsPatch []*model.ChannelModerationPatch) ([]*model.ChannelModeration, *model.AppError) {
	ret := _m.Called(channel, channelModerationsPatch)

	var r0 []*model.ChannelModeration
	if rf, ok := ret.Get(0).(func(*model.Channel, []*model.ChannelModerationPatch) []*model.ChannelModeration); ok {
		r0 = rf(channel, channelModerationsPatch)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.ChannelModeration)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(*model.Channel, []*model.ChannelModerationPatch) *model.AppError); ok {
		r1 = rf(channel, channelModerationsPatch)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// PermanentDeleteChannel provides a mock function with given fields: channel
func (_m *MockAppIface) PermanentDeleteChannel(channel *model.Channel) *model.AppError {
	ret := _m.Called(channel)

	var r0 *model.AppError
	if rf, ok := ret.Get(0).(func(*model.Channel) *model.AppError); ok {
		r0 = rf(channel)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AppError)
		}
	}

	return r0
}

// SaveReactionForPost provides a mock function with given fields: c, reaction
func (_m *MockAppIface) SaveReactionForPost(c *request.Context, reaction *model.Reaction) (*model.Reaction, *model.AppError) {
	ret := _m.Called(c, reaction)

	var r0 *model.Reaction
	if rf, ok := ret.Get(0).(func(*request.Context, *model.Reaction) *model.Reaction); ok {
		r0 = rf(c, reaction)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Reaction)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(*request.Context, *model.Reaction) *model.AppError); ok {
		r1 = rf(c, reaction)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}

// SendEphemeralPost provides a mock function with given fields: userId, post
func (_m *MockAppIface) SendEphemeralPost(userId string, post *model.Post) *model.Post {
	ret := _m.Called(userId, post)

	var r0 *model.Post
	if rf, ok := ret.Get(0).(func(string, *model.Post) *model.Post); ok {
		r0 = rf(userId, post)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Post)
		}
	}

	return r0
}

// UpdatePost provides a mock function with given fields: c, post, safeUpdate
func (_m *MockAppIface) UpdatePost(c *request.Context, post *model.Post, safeUpdate bool) (*model.Post, *model.AppError) {
	ret := _m.Called(c, post, safeUpdate)

	var r0 *model.Post
	if rf, ok := ret.Get(0).(func(*request.Context, *model.Post, bool) *model.Post); ok {
		r0 = rf(c, post, safeUpdate)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Post)
		}
	}

	var r1 *model.AppError
	if rf, ok := ret.Get(1).(func(*request.Context, *model.Post, bool) *model.AppError); ok {
		r1 = rf(c, post, safeUpdate)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.AppError)
		}
	}

	return r0, r1
}
