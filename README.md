<h1 align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://assets.raito.io/icons/logo-vertical-dark@2x.png">
    <img height="250px" src="https://assets.raito.io/icons/logo-vertical@2x.png">
  </picture>
</h1>

<h4 align="center">
  Azure Active Directory plugin for the Raito CLI
</h4>

<p align="center">
    <a href="/LICENSE.md" target="_blank"><img src="https://img.shields.io/badge/license-Apache%202-brightgreen.svg" alt="Software License" /></a>
    <a href="https://github.com/raito-io/cli-plugin-azure-ad/actions/workflows/build.yml" target="_blank"><img src="https://img.shields.io/github/workflow/status/raito-io/cli-plugin-azure-ad/Raito%20CLI%20-%Azure%20Active%20Directory%20Plugin%20-%20Build/main" alt="Build status" /></a>
    <a href="https://codecov.io/gh/raito-io/cli-plugin-azure-ad" target="_blank"><img src="https://img.shields.io/codecov/c/github/raito-io/cli-plugin-azure-ad" alt="Code Coverage" /></a>
</p>

<hr/>

# Raito CLI Plugin - Azure Active Directory

This Raito CLI plugin implements the integration with Azure Active Directory. It can
TODO


## Prerequisites
To use this plugin, you will need

1. The Raito CLI to be correctly installed. You can check out our [documentation](http://docs.raito.io/docs/cli/installation) for help on this.
2. A Raito Cloud account to synchronize your Azure Active Directory with. If you don't have this yet, visit our webpage at (https://raito.io) and request a trial account.
3. At least one Azure account with Active Directory setup

## Usage
To use the plugin, add the following snippet to your Raito CLI configuration file (`raito.yml`, by default) under the `targets` section:

```json
  - name: azure-ad1
    connector-name: raito-io/cli-plugin-azure-ad
    identity-store-id: <<Active Directory IdentityStore ID>>

    # Specifying the Azure Active Directory specific config parameters
    TODO
```

Next, replace the values of the indicated fields with your specific values:
TODO

You will also need to configure the Raito CLI further to connect to your Raito Cloud account, if that's not set up yet.
A full guide on how to configure the Raito CLI can be found on (http://docs.raito.io/docs/cli/configuration).

## Trying it out

As a first step, you can check if the CLI finds this plugin correctly. In a command-line terminal, execute the following command:
```bash
$> raito info raito-io/cli-plugin-azure-ad
```

This will download the latest version of the plugin (if you don't have it yet) and output the name and version of the plugin, together with all the plugin-specific parameters to configure it.

When you are ready to try out the synchronization for the first time, execute:
```bash
$> raito run
```
This will take the configuration from the `raito.yml` file (in the current working directory) and start a single synchronization.

Note: if you have multiple targets configured in your configuration file, you can run only this target by adding `--only-targets azure-ad1` at the end of the command.
