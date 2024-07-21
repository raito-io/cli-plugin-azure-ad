//go:build integration

package it

import (
	"context"
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
