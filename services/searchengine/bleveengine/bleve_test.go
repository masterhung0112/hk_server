// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package bleveengine

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/blevesearch/bleve"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/services/searchengine"
	"github.com/masterhung0112/hk_server/store/searchlayer"
	"github.com/masterhung0112/hk_server/store/searchtest"
	"github.com/masterhung0112/hk_server/store/sqlstore"
	"github.com/masterhung0112/hk_server/store/storetest"
	"github.com/masterhung0112/hk_server/testlib"
)

type BleveEngineTestSuite struct {
	suite.Suite

	SQLSettings  *model.SqlSettings
	SQLStore     *sqlstore.SqlStore
	SearchEngine *searchengine.Broker
	Store        *searchlayer.SearchStore
	BleveEngine  *BleveEngine
	IndexDir     string
}

func TestBleveEngineTestSuite(t *testing.T) {
	suite.Run(t, new(BleveEngineTestSuite))
}

func (s *BleveEngineTestSuite) setupIndexes() {
	indexDir, err := ioutil.TempDir("", "mmbleve")
	if err != nil {
		s.Require().FailNow("Cannot setup bleveengine tests: %s", err.Error())
	}
	s.IndexDir = indexDir
}

func (s *BleveEngineTestSuite) setupStore() {
	driverName := os.Getenv("MM_SQLSETTINGS_DRIVERNAME")
	if driverName == "" {
		driverName = model.DATABASE_DRIVER_POSTGRES
	}
	s.SQLSettings = storetest.MakeSqlSettings(driverName)
	s.SQLStore = sqlstore.New(*s.SQLSettings, nil)

	cfg := &model.Config{}
	cfg.SetDefaults()
	cfg.BleveSettings.EnableIndexing = model.NewBool(true)
	cfg.BleveSettings.EnableSearching = model.NewBool(true)
	cfg.BleveSettings.EnableAutocomplete = model.NewBool(true)
	cfg.BleveSettings.IndexDir = model.NewString(s.IndexDir)
	cfg.SqlSettings.DisableDatabaseSearch = model.NewBool(true)

	s.SearchEngine = searchengine.NewBroker(cfg, nil)
	s.Store = searchlayer.NewSearchLayer(&testlib.TestStore{Store: s.SQLStore}, s.SearchEngine, cfg)

	s.BleveEngine = NewBleveEngine(cfg, nil)
	s.BleveEngine.indexSync = true
	s.SearchEngine.RegisterBleveEngine(s.BleveEngine)
	if err := s.BleveEngine.Start(); err != nil {
		s.Require().FailNow("Cannot start bleveengine: %s", err.Error())
	}
}

func (s *BleveEngineTestSuite) SetupSuite() {
	s.setupIndexes()
	s.setupStore()
}

func (s *BleveEngineTestSuite) TearDownSuite() {
	os.RemoveAll(s.IndexDir)
	s.SQLStore.Close()
	storetest.CleanupSqlSettings(s.SQLSettings)
}

func (s *BleveEngineTestSuite) TestBleveSearchStoreTests() {
	searchTestEngine := &searchtest.SearchTestEngine{
		Driver: searchtest.ENGINE_BLEVE,
	}

	s.Run("TestSearchChannelStore", func() {
		searchtest.TestSearchChannelStore(s.T(), s.Store, searchTestEngine)
	})

	s.Run("TestSearchUserStore", func() {
		searchtest.TestSearchUserStore(s.T(), s.Store, searchTestEngine)
	})

	s.Run("TestSearchPostStore", func() {
		searchtest.TestSearchPostStore(s.T(), s.Store, searchTestEngine)
	})
}

func (s *BleveEngineTestSuite) TestDeleteChannelPosts() {
	s.Run("Should remove all the posts that belongs to a channel", func() {
		s.BleveEngine.PurgeIndexes()
		teamID := model.NewId()
		userID := model.NewId()
		channelID := model.NewId()
		channelToAvoidID := model.NewId()
		posts := make([]*model.Post, 0)
		for i := 0; i < 10; i++ {
			post := createPost(userID, channelID, "test one two three")
			appErr := s.SearchEngine.BleveEngine.IndexPost(post, teamID)
			require.Nil(s.T(), appErr)
			posts = append(posts, post)
		}
		postToAvoid := createPost(userID, channelToAvoidID, "test one two three")
		appErr := s.SearchEngine.BleveEngine.IndexPost(postToAvoid, teamID)
		require.Nil(s.T(), appErr)

		s.SearchEngine.BleveEngine.DeleteChannelPosts(channelID)

		doc, err := s.BleveEngine.PostIndex.Document(postToAvoid.Id)
		require.Nil(s.T(), err)
		require.Equal(s.T(), postToAvoid.Id, doc.ID)
		numberDocs, err := s.BleveEngine.PostIndex.DocCount()
		require.Nil(s.T(), err)
		require.Equal(s.T(), 1, int(numberDocs))
	})

	s.Run("Shouldn't do anything if there is not posts for the selected channel", func() {
		s.BleveEngine.PurgeIndexes()
		teamID := model.NewId()
		userID := model.NewId()
		channelID := model.NewId()
		channelToDeleteID := model.NewId()
		post := createPost(userID, channelID, "test one two three")
		appErr := s.SearchEngine.BleveEngine.IndexPost(post, teamID)
		require.Nil(s.T(), appErr)

		s.SearchEngine.BleveEngine.DeleteChannelPosts(channelToDeleteID)

		_, err := s.BleveEngine.PostIndex.Document(post.Id)
		require.Nil(s.T(), err)
		numberDocs, err := s.BleveEngine.PostIndex.DocCount()
		require.Nil(s.T(), err)
		require.Equal(s.T(), 1, int(numberDocs))
	})
}

func (s *BleveEngineTestSuite) TestDeleteUserPosts() {
	s.Run("Should remove all the posts that belongs to a user", func() {
		s.BleveEngine.PurgeIndexes()
		teamID := model.NewId()
		userID := model.NewId()
		userToAvoidID := model.NewId()
		channelID := model.NewId()
		posts := make([]*model.Post, 0)
		for i := 0; i < 10; i++ {
			post := createPost(userID, channelID, "test one two three")
			appErr := s.SearchEngine.BleveEngine.IndexPost(post, teamID)
			require.Nil(s.T(), appErr)
			posts = append(posts, post)
		}
		postToAvoid := createPost(userToAvoidID, channelID, "test one two three")
		appErr := s.SearchEngine.BleveEngine.IndexPost(postToAvoid, teamID)
		require.Nil(s.T(), appErr)

		s.SearchEngine.BleveEngine.DeleteUserPosts(userID)

		doc, err := s.BleveEngine.PostIndex.Document(postToAvoid.Id)
		require.Nil(s.T(), err)
		require.Equal(s.T(), postToAvoid.Id, doc.ID)
		numberDocs, err := s.BleveEngine.PostIndex.DocCount()
		require.Nil(s.T(), err)
		require.Equal(s.T(), 1, int(numberDocs))
	})

	s.Run("Shouldn't do anything if there is not posts for the selected user", func() {
		s.BleveEngine.PurgeIndexes()
		teamID := model.NewId()
		userID := model.NewId()
		userToDeleteID := model.NewId()
		channelID := model.NewId()
		post := createPost(userID, channelID, "test one two three")
		appErr := s.SearchEngine.BleveEngine.IndexPost(post, teamID)
		require.Nil(s.T(), appErr)

		s.SearchEngine.BleveEngine.DeleteUserPosts(userToDeleteID)

		_, err := s.BleveEngine.PostIndex.Document(post.Id)
		require.Nil(s.T(), err)
		numberDocs, err := s.BleveEngine.PostIndex.DocCount()
		require.Nil(s.T(), err)
		require.Equal(s.T(), 1, int(numberDocs))
	})
}

func (s *BleveEngineTestSuite) TestDeletePosts() {
	s.BleveEngine.PurgeIndexes()
	teamID := model.NewId()
	userID := model.NewId()
	userToAvoidID := model.NewId()
	channelID := model.NewId()
	posts := make([]*model.Post, 0)
	for i := 0; i < 10; i++ {
		post := createPost(userID, channelID, "test one two three")
		appErr := s.SearchEngine.BleveEngine.IndexPost(post, teamID)
		require.Nil(s.T(), appErr)
		posts = append(posts, post)
	}
	postToAvoid := createPost(userToAvoidID, channelID, "test one two three")
	appErr := s.SearchEngine.BleveEngine.IndexPost(postToAvoid, teamID)
	require.Nil(s.T(), appErr)

	query := bleve.NewTermQuery(userID)
	query.SetField("UserId")
	search := bleve.NewSearchRequest(query)
	count, err := s.BleveEngine.deletePosts(search, 1)
	require.Nil(s.T(), err)
	require.Equal(s.T(), 10, int(count))

	doc, err := s.BleveEngine.PostIndex.Document(postToAvoid.Id)
	require.Nil(s.T(), err)
	require.Equal(s.T(), postToAvoid.Id, doc.ID)
	numberDocs, err := s.BleveEngine.PostIndex.DocCount()
	require.Nil(s.T(), err)
	require.Equal(s.T(), 1, int(numberDocs))
}