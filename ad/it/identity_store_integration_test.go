//go:build integration

package it

import (
	"context"
	"sort"
	"testing"

	"github.com/raito-io/cli/base/wrappers/mocks"
	"github.com/stretchr/testify/suite"

	"github.com/raito-io/cli-plugin-azure-ad/ad"
)

type IdentityStoreTestSuite struct {
	AzureAdTestSuite
}

func TestIdentityStoreTestSuite(t *testing.T) {
	ts := IdentityStoreTestSuite{}
	suite.Run(t, &ts)
}

func (s *IdentityStoreTestSuite) TestIdentityStoreSync() {
	//Given
	identityHandler := mocks.NewSimpleIdentityStoreIdentityHandler(s.T(), 27)
	identityStoreSyncer := ad.NewIdentityStoreSyncer()

	//When
	err := identityStoreSyncer.SyncIdentityStore(context.Background(), identityHandler, s.getConfig())

	//Then
	s.NoError(err)

	s.True(len(identityHandler.Users) > 3)
	s.True(len(identityHandler.Groups) >= 3)

	group := ""

	for _, user := range identityHandler.Users {
		if user.Email == "c_harris@raito.io" {
			s.Equal(1, len(user.GroupExternalIds))
			group = user.GroupExternalIds[0]

			s.Len(user.Tags, 2)
			for _, tag := range user.Tags {
				if tag.Key == "Department" {
					s.Equal("Marketing", tag.Value)
				} else if tag.Key == "JobTitle" {
					s.Equal("Data Governance Lead", tag.Value)
				} else {
					s.Fail("Unexpected tag")
				}
			}
			break
		}
	}

	s.NotEqual("", group)

	found := false
	for _, g := range identityHandler.Groups {
		if g.ExternalId == group {
			found = true
			s.Equal("Marketing", g.Name)
		}
	}

	s.True(found)
}

func (s *IdentityStoreTestSuite) TestIdentityStoreSync_GroupFilter() {
	//Given
	identityHandler := mocks.NewSimpleIdentityStoreIdentityHandler(s.T(), 27)
	identityStoreSyncer := ad.NewIdentityStoreSyncer()

	//When
	configMap := s.getConfig()
	configMap.Parameters[ad.AdGroupsFilter] = "08025b48-4886-42b3-a3ea-d772a4267f8d" // Engineering group
	err := identityStoreSyncer.SyncIdentityStore(context.Background(), identityHandler, configMap)

	//Then
	s.NoError(err)

	s.True(len(identityHandler.Users) < 4)
	s.Len(identityHandler.Groups, 2)

	var groupNames []string
	var eGroupId, deGroupId string

	for _, group := range identityHandler.Groups {
		groupNames = append(groupNames, group.Name)

		if group.Name == "Data Engineering" {
			deGroupId = group.ExternalId
			s.Len(group.ParentGroupExternalIds, 1)
		} else {
			eGroupId = group.ExternalId
			s.Len(group.ParentGroupExternalIds, 0)
		}
	}

	sort.Strings(groupNames)
	s.ElementsMatch([]string{"Data Engineering", "Engineering"}, groupNames)

	var userEmails []string

	for _, user := range identityHandler.Users {
		userEmails = append(userEmails, user.Email)
		s.Len(user.GroupExternalIds, 1)

		if user.Email == "b_stewart@raito.io" {
			s.Equal(deGroupId, user.GroupExternalIds[0])
		} else {
			s.Equal(eGroupId, user.GroupExternalIds[0])
		}
	}

	sort.Strings(userEmails)
	s.ElementsMatch([]string{"b_stewart@raito.io", "gill.bates@raito.io", "n_nguyen@raito.io"}, userEmails)
}
