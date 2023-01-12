//go:build integration

package it

import (
	"context"
	"testing"

	"github.com/raito-io/cli-plugin-azure-ad/ad"
	"github.com/raito-io/cli/base/wrappers/mocks"
	"github.com/stretchr/testify/suite"
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
	identityHandler := mocks.NewSimpleIdentityStoreIdentityHandler(s.T(), 1)
	identityStoreSyncer := ad.NewIdentityStoreSyncer()

	//When
	err := identityStoreSyncer.SyncIdentityStore(context.Background(), identityHandler, s.getConfig())

	//Then
	s.NoError(err)

	s.True(len(identityHandler.Users) >= 3)

	s.Empty(identityHandler.Groups)
}
