<h1 align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://github.com/raito-io/raito-io.github.io/raw/master/assets/images/logo-vertical-dark%402x.png">
    <img height="250px" src="https://github.com/raito-io/raito-io.github.io/raw/master/assets/images/logo-vertical%402x.png">
  </picture>
</h1>

<h4 align="center">
  Azure Active Directory plugin for the Raito CLI
</h4>

<p align="center">
    <a href="/LICENSE.md" target="_blank"><img src="https://img.shields.io/badge/license-Apache%202-brightgreen.svg" alt="Software License" /></a>
    <a href="https://github.com/raito-io/cli-plugin-azure-ad/actions/workflows/build.yml" target="_blank"><img src="https://img.shields.io/github/actions/workflow/status/raito-io/cli-plugin-azure-ad/build.yml?branch=main" alt="Build status" /></a>
    <!--a href="https://codecov.io/gh/raito-io/cli-plugin-azure-ad" target="_blank"><img src="https://img.shields.io/codecov/c/github/raito-io/cli-plugin-azure-ad" alt="Code Coverage" /></a-->
</p>

<hr/>

# Raito CLI Plugin - Azure Active Directory

This Raito CLI plugin will synchronize the users and groups from an Azure Active Directory account to a specified Raito Identity Store.


## Prerequisites
To use this plugin, you will need

1. The Raito CLI to be correctly installed. You can check out our [documentation](http://docs.raito.io/docs/cli/installation) for help on this.
2. A Raito Cloud account to synchronize your Azure Active Directory with. If you don't have this yet, visit our webpage at (https://raito.io) and request a trial account.
3. At least one Azure account with Active Directory setup
   1. You'll need the Tenant ID of your directory
   2. Under 'App registrations', set up a new application for this integration. You'll need the Application (client) ID.
   3. In the newly created application go to 'Certificates & secrets' to create a new client secret 
   4. In the newly created application go to 'API Permissions' and make sure the application has the following permissions: `Group.Read.All`, `User.Read` and `User.Read.All`. Make sure these are approved and are of type `Application`. To do this, go to the 'Enterprise applications' menu in your AD Directory, find your application and go to 'Permissions'. 

## Usage
To use the plugin, add the following snippet to your Raito CLI configuration file (`raito.yml`, by default) under the `targets` section:

```json
  - name: azure-ad1
    connector-name: raito-io/cli-plugin-azure-ad
    identity-store-id: <<Active Directory IdentityStore ID>>

    ad-tenantid: <<Your AD Tentant ID>>
    ad-clientid: <<Your AD Client ID>>
    ad-secret: "{{RAITO_AD_SECRET}}"
```

Next, replace the values of the indicated fields with your specific values:
- `<<Your AD Tentant ID>>`: the tenant ID as explained in the prerequisites above
- `<<Your AD Client ID>>`: the Application (client) ID as explained in the prerequisites above

Make sure you have a system variable called `RAITO_AD_SECRET` with the client secret (see prerequisites above) as its value.

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
