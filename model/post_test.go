package model

import (
	"io/ioutil"
	"strings"
	"sync"
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