//go:build integration

package it

import (
	"github.com/raito-io/cli-plugin-azure-ad/ad"
	"os"
	"sync"

	"github.com/raito-io/cli/base/util/config"
	"github.com/stretchr/testify/suite"
)

var (
	adTenantId string
	adClientId string
	adSecret   string
	lock       = &sync.Mutex{}
)

func readDatabaseConfig() *config.ConfigMap {
	lock.Lock()
	defer lock.Unlock()

	if adClientId == "" {
		adClientId = os.Getenv("AD_CLIENTID")
		adTenantId = os.Getenv("AD_TENANTID")
		adSecret = os.Getenv("AD_SECRET")
	}

	return &config.ConfigMap{
		Parameters: map[string]interface{}{
			ad.AdTenantId: adTenantId,
			ad.AdClientId: adClientId,
			ad.AdSecret:   adSecret,
		},
	}
}

type AzureAdTestSuite struct {
	suite.Suite
}

func (s *AzureAdTestSuite) getConfig() *config.ConfigMap {
	return readDatabaseConfig()
}
