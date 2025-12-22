package proxmox

import (
	"bytes"
	"context"
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/Telmate/proxmox-api-go/internal/util"
)

type (
	ApiTokenInterface interface {
		Create(context.Context, UserID, ApiTokenConfig) (ApiTokenSecret, error)
		CreateNoCheck(context.Context, UserID, ApiTokenConfig) (ApiTokenSecret, error)

		// Returns true if the token existed and was deleted, false if the token did not exist.
		Delete(context.Context, ApiTokenID) (deleted bool, err error)
		DeleteNoCheck(context.Context, ApiTokenID) (deleted bool, err error)

		Exists(context.Context, ApiTokenID) (bool, error)
		ExistsNoCheck(context.Context, ApiTokenID) (bool, error)

		List(context.Context, UserID) (RawApiTokens, error)
		ListNoCheck(context.Context, UserID) (RawApiTokens, error)

		Read(context.Context, ApiTokenID) (RawApiTokenConfig, error)
		ReadNoCheck(context.Context, ApiTokenID) (RawApiTokenConfig, error)

		Update(context.Context, UserID, ApiTokenConfig) error
		UpdateNoCheck(context.Context, UserID, ApiTokenConfig) error
	}
	apiTokenClient struct {
		api       *clientAPI
		oldClient *Client
	}
)

var _ ApiTokenInterface = (*apiTokenClient)(nil)

func (c *apiTokenClient) Create(ctx context.Context, userID UserID, token ApiTokenConfig) (ApiTokenSecret, error) {
	if err := userID.Validate(); err != nil {
		return "", err
	}
	if err := token.Validate(); err != nil {
		return "", err
	}
	return c.CreateNoCheck(ctx, userID, token)
}

func (c *apiTokenClient) CreateNoCheck(ctx context.Context, userID UserID, token ApiTokenConfig) (ApiTokenSecret, error) {
	return token.create(ctx, c.api, userID)
}

func (c *apiTokenClient) Delete(ctx context.Context, token ApiTokenID) (bool, error) {
	if err := token.Validate(); err != nil {
		return false, err
	}
	return c.DeleteNoCheck(ctx, token)
}

func (c *apiTokenClient) DeleteNoCheck(ctx context.Context, token ApiTokenID) (bool, error) {
	return token.delete(ctx, c.api)
}

func (c *apiTokenClient) Exists(ctx context.Context, token ApiTokenID) (bool, error) {
	if err := token.Validate(); err != nil {
		return false, err
	}
	return c.ExistsNoCheck(ctx, token)
}

func (c *apiTokenClient) ExistsNoCheck(ctx context.Context, token ApiTokenID) (bool, error) {
	return token.exists(ctx, c.api)
}

func (c *apiTokenClient) List(ctx context.Context, userID UserID) (RawApiTokens, error) {
	if err := userID.Validate(); err != nil {
		return nil, err
	}
	return c.ListNoCheck(ctx, userID)
}

func (c *apiTokenClient) ListNoCheck(ctx context.Context, userID UserID) (RawApiTokens, error) {
	return userID.listApiTokens(ctx, c.api)
}

func (c *apiTokenClient) Read(ctx context.Context, tokenID ApiTokenID) (RawApiTokenConfig, error) {
	if err := tokenID.Validate(); err != nil {
		return nil, err
	}
	return c.ReadNoCheck(ctx, tokenID)
}

func (c *apiTokenClient) ReadNoCheck(ctx context.Context, tokenID ApiTokenID) (RawApiTokenConfig, error) {
	token, exists, err := tokenID.read(ctx, c.api)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("api token does not exist")
	}
	return token, nil
}

func (c *apiTokenClient) Update(ctx context.Context, user UserID, token ApiTokenConfig) error {
	if err := user.Validate(); err != nil {
		return err
	}
	if err := token.Validate(); err != nil {
		return err
	}
	return c.UpdateNoCheck(ctx, user, token)
}

func (c *apiTokenClient) UpdateNoCheck(ctx context.Context, user UserID, token ApiTokenConfig) error {
	return token.update(ctx, c.api, user)
}

type ApiTokenConfig struct {
	Name                ApiTokenName `json:"id"`
	Comment             *string      `json:"comment,omitempty"` // Never nil when returned
	Expiration          *uint        `json:"expire,omitempty"`  // Never nil when returned
	PrivilegeSeparation *bool        `json:"privsep,omitempty"` // Never nil when returned
}

func (token ApiTokenConfig) create(ctx context.Context, c *clientAPI, userID UserID) (ApiTokenSecret, error) {
	params, err := c.postMap(ctx, "/access/users/"+userID.String()+"/token/"+token.Name.String(), token.mapToApiCreate(), "ApiToken", "config", nil)
	if err != nil {
		return "", err
	}
	return ApiTokenSecret(params[apiTokenApiKeySecret].(string)), nil
}

func (token ApiTokenConfig) mapToApiCreate() *[]byte {
	builder := strings.Builder{}
	if token.PrivilegeSeparation != nil {
		builder.WriteString("&" + apiTokenApiKeyPrivilegeSeparation + "=" + boolToIntString(*token.PrivilegeSeparation))
	}
	if token.Expiration != nil && *token.Expiration > 0 {
		builder.WriteString("&" + apiTokenApiKeyExpiration + "=" + strconv.FormatUint(uint64(*token.Expiration), 10))
	}
	if token.Comment != nil && *token.Comment != "" {
		builder.WriteString("&" + apiTokenApiKeyComment + "=" + url.QueryEscape(*token.Comment))
	}
	if builder.Len() > 0 {
		body := bytes.NewBufferString(builder.String()[1:]).Bytes()
		return &body
	}
	return nil
}

func (token ApiTokenConfig) mapToApiUpdate() *[]byte {
	builder := strings.Builder{}
	if token.PrivilegeSeparation != nil {
		builder.WriteString("&" + apiTokenApiKeyPrivilegeSeparation + "=" + boolToIntString(*token.PrivilegeSeparation))
	}
	if token.Expiration != nil {
		builder.WriteString("&" + apiTokenApiKeyExpiration + "=" + strconv.FormatUint(uint64(*token.Expiration), 10))
	}
	if token.Comment != nil {
		builder.WriteString("&" + apiTokenApiKeyComment + "=" + url.QueryEscape(*token.Comment))
	}
	if builder.Len() > 0 {
		body := bytes.NewBufferString(builder.String()[1:]).Bytes()
		return &body
	}
	return nil
}

func (token ApiTokenConfig) update(ctx context.Context, c *clientAPI, userID UserID) error {
	body := token.mapToApiUpdate()
	if body == nil {
		return nil
	}
	_, err := c.postRawRetry(ctx, "/access/users/"+userID.String()+"/token/"+token.Name.String(), body, 3, nil)
	return err
}

func (token ApiTokenConfig) Validate() error { return token.Name.Validate() }

type (
	RawApiTokens interface {
		FormatArray() []RawApiTokenConfig
		FormatMap() map[ApiTokenName]RawApiTokenConfig
		Len() int
		SelectName(ApiTokenName) (RawApiTokenConfig, bool)
	}
	rawApiTokens struct{ a []any }
)

var _ RawApiTokens = (*rawApiTokens)(nil)

func (raw *rawApiTokens) SelectName(name ApiTokenName) (RawApiTokenConfig, bool) {
	for i := range raw.a {
		tmpMap := raw.a[i].(map[string]any)
		if v, ok := tmpMap[apiTokenApiKeyTokenID]; ok {
			if v.(string) == name.String() {
				return &rawApiTokenConfig{a: tmpMap, name: name}, true
			}
		}
	}
	return nil, false
}

func (raw *rawApiTokens) Len() int { return len(raw.a) }

func (raw *rawApiTokens) FormatArray() []RawApiTokenConfig {
	tokenArray := make([]RawApiTokenConfig, len(raw.a))
	for i := range raw.a {
		tmpMap := raw.a[i].(map[string]any)
		var name ApiTokenName
		if v, ok := tmpMap[apiTokenApiKeyTokenID]; ok {
			name = ApiTokenName(v.(string))
		}
		tokenArray[i] = &rawApiTokenConfig{a: tmpMap, name: name}
	}
	return tokenArray
}

func (raw *rawApiTokens) FormatMap() map[ApiTokenName]RawApiTokenConfig {
	tokenMap := make(map[ApiTokenName]RawApiTokenConfig, len(raw.a))
	for i := range raw.a {
		tmpMap := raw.a[i].(map[string]any)
		var name ApiTokenName
		if v, ok := tmpMap[apiTokenApiKeyTokenID]; ok {
			name = ApiTokenName(v.(string))
		}
		tokenMap[name] = &rawApiTokenConfig{a: tmpMap, name: name}
	}
	return tokenMap
}

type (
	RawApiTokenConfig interface {
		Get() ApiTokenConfig
		GetName() ApiTokenName
		GetComment() string
		GetExpiration() uint
		GetPrivilegeSeparation() bool
	}
	rawApiTokenConfig struct {
		a    map[string]any
		name ApiTokenName
	}
)

var _ RawApiTokenConfig = (*rawApiTokenConfig)(nil)

func (raw *rawApiTokenConfig) Get() ApiTokenConfig {
	return ApiTokenConfig{
		Comment:             util.Pointer(raw.GetComment()),
		Expiration:          util.Pointer(raw.GetExpiration()),
		Name:                raw.GetName(),
		PrivilegeSeparation: util.Pointer(raw.GetPrivilegeSeparation())}
}

func (raw *rawApiTokenConfig) GetName() ApiTokenName { return raw.name }

func (raw *rawApiTokenConfig) GetComment() string {
	if v, ok := raw.a[apiTokenApiKeyComment]; ok {
		return v.(string)
	}
	return ""
}

func (raw *rawApiTokenConfig) GetExpiration() uint {
	if v, ok := raw.a[apiTokenApiKeyExpiration]; ok {
		return uint(v.(float64))
	}
	return 0
}

func (raw *rawApiTokenConfig) GetPrivilegeSeparation() bool {
	if v, ok := raw.a[apiTokenApiKeyPrivilegeSeparation]; ok {
		return int(v.(float64)) == 1
	}
	return false
}

type ApiTokenID struct {
	User      UserID
	TokenName ApiTokenName
}

func (id ApiTokenID) delete(ctx context.Context, c *clientAPI) (bool, error) {
	err := c.deleteRetry(ctx, "/access/users/"+id.User.String()+"/token/"+id.TokenName.String(), 3, nil)
	var apiErr *ApiError
	if errors.As(err, &apiErr) {
		if strings.HasPrefix(apiErr.Message, "no such token ") {
			return false, nil
		}
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (id ApiTokenID) exists(ctx context.Context, c *clientAPI) (bool, error) {
	_, exists, err := id.read(ctx, c)
	return exists, err
}

const ApiTokenID_Errors_Invalid string = "api token ID must be in the format user@realm!tokenname"

func (id *ApiTokenID) Parse(s string) error {
	indexAt := strings.IndexRune(s, '@')
	if indexAt == -1 || indexAt == 0 || indexAt == len(s)-1 {
		return errors.New(ApiTokenID_Errors_Invalid)
	}
	indexEx := strings.IndexRune(s[indexAt+1:], '!')
	if indexEx == -1 || indexEx == 0 || indexEx == len(s[indexAt+1:])-1 {
		return errors.New(ApiTokenID_Errors_Invalid)
	}
	id.User.Name = s[:indexAt]
	id.User.Realm = s[indexAt+1 : indexAt+1+indexEx]
	id.TokenName = ApiTokenName(s[indexAt+1+indexEx+1:])
	return nil
}

func (id ApiTokenID) read(ctx context.Context, c *clientAPI) (*rawApiTokenConfig, bool, error) {
	data, err := c.getMap(ctx, "/access/users/"+id.User.String()+"/token/"+id.TokenName.String(), "api token", "CONFIG", nil)
	if err != nil {
		var apiErr *ApiError
		if errors.As(err, &apiErr) {
			if strings.HasPrefix(apiErr.Message, "no such token ") {
				return nil, false, nil
			}
		}
		return nil, false, err
	}
	return &rawApiTokenConfig{a: data, name: id.TokenName}, true, nil
}

func (id ApiTokenID) String() string { return id.User.String() + "!" + id.TokenName.String() } // Used for fmt.Stringer interface

func (id ApiTokenID) Validate() error {
	if err := id.User.Validate(); err != nil {
		return err
	}
	if err := id.TokenName.Validate(); err != nil {
		return err
	}
	return nil
}

// ^[A-Za-z][A-Za-z0-9\.\-_]{1,127}
type ApiTokenName string

var regexApiTokenName = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9\.\-_]+$`)

const ApiTokenName_Errors_Invalid string = `api token name must match the following regex: ^[A-Za-z][A-Za-z0-9\.-_]{1,127}$`

func (name ApiTokenName) String() string { return string(name) } // Used for fmt.Stringer interface

func (name ApiTokenName) Validate() error {
	if len(name) == 0 || len(name) > 128 {
		return errors.New(ApiTokenName_Errors_Invalid)
	}
	if !regexApiTokenName.MatchString(name.String()) {
		return errors.New(ApiTokenName_Errors_Invalid)
	}
	return nil
}

type ApiTokenSecret string

func (s ApiTokenSecret) String() string { return string(s) } // Used for fmt.Stringer interface

const (
	apiTokenApiKeyTokenID             string = "tokenid"
	apiTokenApiKeyComment             string = "comment"
	apiTokenApiKeyExpiration          string = "expire"
	apiTokenApiKeyPrivilegeSeparation string = "privsep"
	apiTokenApiKeySecret              string = "value"
)
