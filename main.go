package main

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/raito-io/cli-plugin-azure-ad/ad"
	"github.com/raito-io/cli/base"
	"github.com/raito-io/cli/base/info"
	"github.com/raito-io/cli/base/util/plugin"
	"github.com/raito-io/cli/base/wrappers"
)

var version = "0.0.0"

var logger hclog.Logger

func main() {
	logger = base.Logger()
	logger.SetLevel(hclog.Debug)

	err := base.RegisterPlugins(wrappers.IdentityStoreSync(ad.NewIdentityStoreSyncer()),
		&info.InfoImpl{
			Info: &plugin.PluginInfo{
				Name:    "Azure Active Directory",
				Version: plugin.ParseVersion(version),
				Parameters: []*plugin.ParameterInfo{
					{Name: "ad-tenantid", Description: "The tenant ID for Azure Active Directory", Mandatory: true},
					{Name: "ad-clientid", Description: "The client ID for Azure Active Directory", Mandatory: true},
					{Name: "ad-secret", Description: "The secret to connect to Azure Active Directory", Mandatory: true},
				},
				Type: []plugin.PluginType{plugin.PluginType_PLUGIN_TYPE_IS_SYNC},
			},
		})

	if err != nil {
		logger.Error(fmt.Sprintf("error while registering plugins: %s", err.Error()))
	}
}
