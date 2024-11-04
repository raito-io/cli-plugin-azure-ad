package ad

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
	is "github.com/raito-io/cli/base/identity_store"
	"github.com/raito-io/cli/base/tag"
	"github.com/raito-io/cli/base/util/config"
	e "github.com/raito-io/cli/base/util/error"
	"github.com/raito-io/cli/base/wrappers"
	"github.com/raito-io/golang-set/set"
)

const startTransitiveMembersURL = "https://graph.microsoft.com/v1.0/groups/{groupId}/transitiveMembers?$select=displayName,jobTitle,mail,officeLocation,userPrincipalName,id,department,description"
const startUsersURL = "https://graph.microsoft.com/v1.0/users/?$select=displayName,jobTitle,mail,officeLocation,userPrincipalName,id,department"
const startGroupsURL = "https://graph.microsoft.com/v1.0/groups/"

const tagSource = "Azure Entra ID"

// This can be made configurable in the future
var userTags = map[string]string{
	"department":     "Department",
	"jobTitle":       "JobTitle",
	"officeLocation": "OfficeLocation",
}

type IdentityStoreSyncer struct {
	parents       map[string]map[string]struct{}
	accessToken   string
	identityCache *IdentityContainer
	handledUsers  set.Set[string]
}

type IdentityContainer struct {
	Users  []*is.User
	Groups []*is.Group
}

func NewIdentityStoreSyncer() *IdentityStoreSyncer {
	return &IdentityStoreSyncer{}
}

func (s *IdentityStoreSyncer) GetIdentityStoreMetaData(_ context.Context, _ *config.ConfigMap) (*is.MetaData, error) {
	logger.Debug("Returning meta data for Azure Active Directory identity store")

	return &is.MetaData{
		Type:        "azure-ad",
		CanBeLinked: true,
		CanBeMaster: true,
	}, nil
}

func (s *IdentityStoreSyncer) SyncIdentityStore(ctx context.Context, identityHandler wrappers.IdentityStoreIdentityHandler, configMap *config.ConfigMap) error {
	s.parents = make(map[string]map[string]struct{})

	container, err := s.GetIdentityContainer(ctx, configMap.Parameters)

	if err != nil {
		return err
	}

	err = identityHandler.AddGroups(container.Groups...)

	if err != nil {
		logger.Error(fmt.Sprintf("error while adding groups: %s", err.Error()))
		return err
	}

	err = identityHandler.AddUsers(container.Users...)

	if err != nil {
		logger.Error(fmt.Sprintf("error while adding users: %s", err.Error()))
		return err
	}

	return nil
}

func (s *IdentityStoreSyncer) GetIdentityContainer(ctx context.Context, params map[string]string) (*IdentityContainer, error) {
	if s.identityCache != nil {
		return s.identityCache, nil
	}

	if s.parents == nil {
		s.parents = make(map[string]map[string]struct{})
	}

	s.identityCache = &IdentityContainer{Users: make([]*is.User, 0), Groups: make([]*is.Group, 0)}

	accessToken, err := s.getAccessToken(params)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch access token: %w", err)
	}

	s.accessToken = accessToken

	logger.Info("Fetched access token for Azure AD")

	groupsFilter := params[AdGroupsFilter]
	strings.TrimSpace(groupsFilter)

	if groupsFilter != "" { // Filtered on groups
		s.handledUsers = set.NewSet[string]()
		groups := strings.Split(groupsFilter, ",")

		for _, group := range groups {
			group = strings.TrimSpace(group)

			// First fetching the group members of this group itself
			err = s.handleGroupMembers(group)
			if err != nil {
				s.identityCache = nil
				return nil, fmt.Errorf("error while processing users in group %q: %w", group, err)
			}

			// Then fetching the group members for all the groups underneath this group
			startURL := strings.Replace(startTransitiveMembersURL, "{groupId}", group, 1)

			err = s.processData(startURL, s.buildFilteredParentsMap) // Groups and users are handled together
			if err != nil {
				s.identityCache = nil
				return nil, fmt.Errorf("error while processing users in group %q: %w", group, err)
			}
		}

		logger.Info(fmt.Sprintf("Built parent map (size %d)", len(s.parents)))

		s.handledUsers = set.NewSet[string]()

		for _, group := range groups {
			group = strings.TrimSpace(group)

			// First handle the group itself.
			err = s.processSingleGroup(group)

			if err != nil {
				s.identityCache = nil
				return nil, fmt.Errorf("error while processing group %q: %w", group, err)
			}

			// Now handling all entities underneath the group
			startURL := strings.Replace(startTransitiveMembersURL, "{groupId}", group, 1)

			err = s.processData(startURL, s.processFilteredEntity) // Groups and users are handled together
			if err != nil {
				s.identityCache = nil
				return nil, fmt.Errorf("error while processing users in group %q: %w", group, err)
			}
		}
	} else { // Or the unfiltered version
		err = s.buildParentsMap()
		if err != nil {
			return nil, err
		}

		logger.Info(fmt.Sprintf("Built parent map (size %d)", len(s.parents)))

		err = s.processData(startGroupsURL, s.processGroup)
		if err != nil {
			s.identityCache = nil
			return nil, fmt.Errorf("error while processing groups: %w", err)
		}

		err = s.processData(startUsersURL, s.processUser)
		if err != nil {
			s.identityCache = nil
			return nil, fmt.Errorf("error while processing users: %w", err)
		}
	}

	return s.identityCache, nil
}

func (s *IdentityStoreSyncer) buildFilteredParentsMap(row map[string]interface{}) error {
	id := row["id"]

	// First check if we didn't already handle this user.
	if s.handledUsers.Contains(id.(string)) {
		return nil
	}

	s.handledUsers.Add(id.(string))

	if row["@odata.type"] == "#microsoft.graph.group" {
		return s.handleGroupMembers(id.(string))
	}

	return nil // Ignoring other types
}

func (s *IdentityStoreSyncer) buildParentsMap() error {
	return s.buildParentsMapForGroups(startGroupsURL + "?$select=id")
}

func (s *IdentityStoreSyncer) buildParentsMapForGroups(url string) error {
	result, err := s.fetchJSONData(url)
	if err != nil {
		return fmt.Errorf("error while fetching the list of groups: %w", err)
	}

	value, valueFound := result["value"]
	if valueFound {
		for _, r := range value.([]interface{}) {
			var row = r.(map[string]interface{})
			id, f := row["id"]

			if f && id != nil {
				err = s.handleGroupMembers(id.(string))
				if err != nil {
					return err
				}
			}
		}
	}

	nextLink, nextFound := result["@odata.nextLink"]
	if nextFound {
		err = s.buildParentsMapForGroups(nextLink.(string))
	}

	return err
}

func (s *IdentityStoreSyncer) handleGroupMembers(parentId string) error {
	members := make([]string, 0, 20)
	members, err := s.getGroupMembers("https://graph.microsoft.com/v1.0/groups/"+parentId+"/members?$select=id", members)

	if err != nil {
		return err
	}

	for _, m := range members {
		mm, f := s.parents[m]
		if !f {
			mm = map[string]struct{}{}
			s.parents[m] = mm
		}

		if _, f := mm[parentId]; !f {
			mm[parentId] = struct{}{}
		}
	}

	return nil
}

func (s *IdentityStoreSyncer) getGroupMembers(url string, members []string) ([]string, error) {
	result, err := s.fetchJSONData(url)
	if err != nil {
		return nil, fmt.Errorf("error while fetching group members: %w", err)
	}

	value, valueFound := result["value"]
	if valueFound {
		for _, r := range value.([]interface{}) {
			var row = r.(map[string]interface{})
			id, f := row["id"]

			if f && id != nil {
				members = append(members, id.(string))
			}
		}
	}

	nextLink, nextFound := result["@odata.nextLink"]
	if nextFound {
		members, err = s.getGroupMembers(nextLink.(string), members)
	}

	return members, err
}

func (s *IdentityStoreSyncer) processGroup(row map[string]interface{}) error {
	id := row["id"]

	logger.Debug(fmt.Sprintf("Processing group %q", id.(string)))

	name := row["displayName"]

	group := is.Group{
		ExternalId:  id.(string),
		Name:        name.(string),
		DisplayName: name.(string),
	}

	if description, ok := row["description"]; ok && description != nil {
		group.Description = description.(string)
	}

	if parents, f := s.parents[id.(string)]; f {
		pl := make([]string, 0, len(parents))
		for parent := range parents {
			pl = append(pl, parent)
		}
		group.ParentGroupExternalIds = pl
	}

	s.identityCache.Groups = append(s.identityCache.Groups, &group)

	return nil
}

func (s *IdentityStoreSyncer) processSingleGroup(group string) error {
	result, err := s.fetchJSONData(startGroupsURL + "/" + group)
	if err != nil {
		return err
	}

	return s.processGroup(result)
}

func (s *IdentityStoreSyncer) processFilteredEntity(row map[string]interface{}) error {
	id := row["id"]

	// First check if we didn't already handle this user.
	if s.handledUsers.Contains(id.(string)) {
		return nil
	}

	s.handledUsers.Add(id.(string))

	if row["@odata.type"] == "#microsoft.graph.user" {
		return s.processUser(row)
	} else if row["@odata.type"] == "#microsoft.graph.group" {
		return s.processGroup(row)
	}

	return nil // Ignoring other types
}

func (s *IdentityStoreSyncer) processUser(row map[string]interface{}) error {
	id := row["id"]

	logger.Debug(fmt.Sprintf("Processing user %q", id.(string)))

	userName := row["userPrincipalName"]
	name := row["displayName"]

	user := is.User{
		ExternalId: id.(string),
		Name:       name.(string),
		UserName:   userName.(string),
	}

	if mail, ok := row["mail"]; ok && mail != nil {
		user.Email = mail.(string)
	} else {
		user.Email = ""
	}

	for field, tagKey := range userTags {
		if value, ok := row[field]; ok && value != nil && value != "" {
			user.Tags = append(user.Tags, &tag.Tag{
				Key:    tagKey,
				Value:  value.(string),
				Source: tagSource,
			})
		}
	}

	if parents, f := s.parents[id.(string)]; f {
		pl := make([]string, 0, len(parents))
		for parent := range parents {
			pl = append(pl, parent)
		}
		user.GroupExternalIds = pl
	}

	s.identityCache.Users = append(s.identityCache.Users, &user)

	return nil
}

func (s *IdentityStoreSyncer) processData(url string, processElement func(map[string]interface{}) error) error {
	result, err := s.fetchJSONData(url)
	if err != nil {
		return err
	}

	nextLink, nextFound := result["@odata.nextLink"]

	value, valueFound := result["value"]
	if valueFound {
		for _, r := range value.([]interface{}) {
			var row = r.(map[string]interface{})

			err = processElement(row)
			if err != nil {
				return fmt.Errorf("error while processing element: %w", err)
			}
		}
	}

	if nextFound {
		err = s.processData(nextLink.(string), processElement)
	}

	return err
}

func (s *IdentityStoreSyncer) getAccessToken(params map[string]string) (string, error) {
	secret := params[AdSecret]

	if secret == "" {
		return "", e.CreateMissingInputParameterError(AdSecret)
	}

	clientId := params[AdClientId]

	if secret == "" {
		return "", e.CreateMissingInputParameterError(AdClientId)
	}

	tenantId := params[AdTenantId]

	if secret == "" {
		return "", e.CreateMissingInputParameterError(AdTenantId)
	}

	// Initializing the client credential
	cred, err := confidential.NewCredFromSecret(secret)
	if err != nil {
		return "", fmt.Errorf("could not create a credential from a secret: %w", err)
	}

	app, err := confidential.New("https://login.microsoftonline.com/"+tenantId, clientId, cred)
	if err != nil {
		return "", fmt.Errorf("could not connect to Microsoft login: %w", err)
	}

	scopes := []string{"https://graph.microsoft.com/.default"}
	result, err := app.AcquireTokenSilent(context.Background(), scopes)

	if err != nil {
		result, err = app.AcquireTokenByCredential(context.Background(), scopes)

		if err != nil {
			return "", fmt.Errorf("unable to fetch access token to access Microsoft Graph: %w", err)
		}
	}

	return result.AccessToken, nil
}

func (s *IdentityStoreSyncer) fetchJSONData(url string) (map[string]interface{}, error) {
	req, _ := http.NewRequest("GET", url, http.NoBody)
	req.Header.Add("Authorization", "Bearer "+s.accessToken)
	client := &http.Client{}

	logger.Info(fmt.Sprintf("Doing GET to %s", url))

	resp, err := client.Do(req)
	if err != nil {
		logger.Error(fmt.Sprintf("error while requesting to url %q: %s", url, err.Error()))
		return nil, err
	}

	body := resp.Body
	defer body.Close()

	byteValue, _ := io.ReadAll(body)

	if logger.IsDebug() {
		logger.Debug(fmt.Sprintf("Got response from: %s", string(byteValue)))
	}

	var result map[string]interface{}
	err = json.Unmarshal(byteValue, &result)

	if err != nil {
		logger.Error(fmt.Sprintf("error while parsing result from url %q: %s ... %s", url, err.Error(), string(byteValue)))
	}

	return result, err
}
