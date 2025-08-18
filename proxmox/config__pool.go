package proxmox

import (
	"context"
	"errors"
	"regexp"
	"strconv"

	"github.com/Telmate/proxmox-api-go/internal/util"
)

func ListPools(ctx context.Context, c *Client) ([]PoolName, error) {
	raw, err := listPools(ctx, c)
	if err != nil {
		return nil, err
	}
	pools := make([]PoolName, len(raw))
	for i, e := range raw {
		pools[i] = PoolName(e.(map[string]interface{})["poolid"].(string))
	}
	return pools, nil
}

func ListPoolsWithComments(ctx context.Context, c *Client) (map[PoolName]string, error) {
	raw, err := listPools(ctx, c)
	if err != nil {
		return nil, err
	}
	pools := make(map[PoolName]string, len(raw))
	for _, e := range raw {
		pool := e.(map[string]interface{})
		var comment string
		if v, isSet := pool["comment"]; isSet {
			comment = v.(string)
		}
		pools[PoolName(pool["poolid"].(string))] = comment
	}
	return pools, nil
}

func listPools(ctx context.Context, c *Client) ([]interface{}, error) {
	if c == nil {
		return nil, errors.New(Client_Error_Nil)
	}
	return c.GetItemListInterfaceArray(ctx, "/pools")
}

type ConfigPool struct {
	Name    PoolName   `json:"name"`
	Comment *string    `json:"comment"`
	Guests  *[]GuestID `json:"guests"`
}

func (config ConfigPool) mapToApi(currentConfig *ConfigPool) map[string]interface{} {
	params := map[string]interface{}{}
	if currentConfig == nil { //create
		params["poolid"] = string(config.Name)
		if config.Comment != nil && *config.Comment != "" {
			params["comment"] = string(*config.Comment)
		}
		return params
	}
	// update
	if config.Comment != nil {
		params["comment"] = string(*config.Comment)
	}
	return params
}

func (ConfigPool) mapToSDK(params map[string]interface{}) (config ConfigPool) {
	if v, isSet := params["poolid"]; isSet {
		config.Name = PoolName(v.(string))
	}
	if v, isSet := params["comment"]; isSet {
		config.Comment = util.Pointer(v.(string))
	}
	if v, isSet := params["members"]; isSet {
		guests := make([]GuestID, 0)
		for _, e := range v.([]interface{}) {
			param := e.(map[string]interface{})
			if v, isSet := param["vmid"]; isSet {
				guests = append(guests, GuestID(v.(float64)))
			}
		}
		if len(guests) > 0 {
			config.Guests = &guests
		}
	}
	return
}

func (config ConfigPool) Create(ctx context.Context, c *Client) error {
	if err := config.Validate(); err != nil {
		return err
	}
	// TODO check permissions
	if exists, err := config.Name.ExistsNoCheck(ctx, c); err != nil {
		return err
	} else if exists {
		return errors.New(PoolName_Error_Exists)
	}
	return config.CreateNoCheck(ctx, c)
}

// CreateNoCheck creates a new pool without validating the input
func (config ConfigPool) CreateNoCheck(ctx context.Context, c *Client) error {
	version, err := c.GetVersion(ctx)
	if err != nil {
		return err
	}
	if err := c.Post(ctx, config.mapToApi(nil), "/pools"); err != nil {
		return err
	}
	if config.Guests != nil {
		return config.Name.addGuestsNoCheck(ctx, c, *config.Guests, nil, version)
	}
	return nil
}

// Same as PoolName.Delete()
func (config ConfigPool) Delete(ctx context.Context, c *Client) error {
	return config.Name.Delete(ctx, c)
}

// Same as PoolName.Exists()
func (config ConfigPool) Exists(ctx context.Context, c *Client) (bool, error) {
	return config.Name.Exists(ctx, c)
}

func (config ConfigPool) Set(ctx context.Context, c *Client) error {
	if err := config.Validate(); err != nil {
		return err
	}
	// TODO check permissions
	return config.SetNoCheck(ctx, c)
}

func (config ConfigPool) SetNoCheck(ctx context.Context, c *Client) error {
	exists, err := config.Name.ExistsNoCheck(ctx, c)
	if err != nil {
		return err
	}
	if exists {
		return config.UpdateNoCheck(ctx, c)
	}
	return config.CreateNoCheck(ctx, c)
}

func (config ConfigPool) Update(ctx context.Context, c *Client) error {
	// TODO add digest during update to check if the config has changed
	if c == nil {
		return errors.New(Client_Error_Nil)
	}
	if err := config.Validate(); err != nil {
		return err
	}
	if exists, err := config.Name.Exists(ctx, c); err != nil {
		return err
	} else if !exists {
		return errors.New(PoolName_Error_NotExists)
	}
	// TODO check permissions
	return config.UpdateNoCheck(ctx, c)
}

// UpdateNoCheck updates a pool without validating the input
func (config ConfigPool) UpdateNoCheck(ctx context.Context, c *Client) error {
	version, err := c.GetVersion(ctx)
	if err != nil {
		return err
	}
	current, err := config.Name.GetNoCheck(ctx, c)
	if err != nil {
		return err
	}
	if params := config.mapToApi(current); len(params) > 0 {
		if err = config.Name.put(ctx, c, params, version); err != nil {
			return err
		}
	}
	if config.Guests != nil {
		return config.Name.SetGuestsNoChecks(ctx, c, *config.Guests)
	}
	return nil
}

func (config ConfigPool) Validate() error {
	// TODO: Add validation for Guests and Comment
	return config.Name.Validate()
}

type PoolName string

const (
	PoolName_Error_Characters        string = "PoolName may only contain the following characters: a-z, A-Z, 0-9, hyphen (-), and underscore (_)"
	PoolName_Error_Empty             string = "PoolName cannot be empty"
	PoolName_Error_Length            string = "PoolName may not be longer than 1024 characters" // proxmox does not seem to have a max length, so we artificially cap it at 1024
	PoolName_Error_NotExists         string = "Pool doesn't exist"
	PoolName_Error_Exists            string = "Pool already exists"
	PoolName_Error_NoGuestsSpecified string = "no guests specified"
)

var regex_PoolName = regexp.MustCompile(`^[a-zA-Z0-9-_]+$`)

func (config PoolName) addGuestsNoCheck(ctx context.Context, c *Client, guestIDs []GuestID, currentGuests *[]GuestID, version Version) error {
	var guestsToAdd []GuestID
	if currentGuests != nil && len(*currentGuests) > 0 {
		guestsToAdd = subtractArray(guestIDs, *currentGuests)
	} else {
		guestsToAdd = guestIDs
	}
	if len(guestsToAdd) == 0 {
		return nil
	}
	if version.Encode() >= version_8_0_0 {
		return config.addGuestsNoCheckV8(ctx, c, guestsToAdd)
	}
	rawGuests, err := listGuests_Unsafe(ctx, c)
	if err != nil {
		return err
	}
	for i, e := range PoolName("").guestsToRemoveFromPools(rawGuests, guestsToAdd) {
		if err = i.removeGuestsNoCheck(ctx, c, e, version); err != nil {
			return err
		}
	}
	return config.addGuestsNoCheckV7(ctx, c, guestsToAdd)
}

func (pool PoolName) addGuestsNoCheckV7(ctx context.Context, c *Client, guestIDs []GuestID) error {
	return pool.putV7(ctx, c, map[string]interface{}{"vms": PoolName("").mapToString(guestIDs)})
}

// from 8.0.0 on proxmox can move the guests to the pool while they are still in another pool
func (pool PoolName) addGuestsNoCheckV8(ctx context.Context, c *Client, guestIDs []GuestID) error {
	return pool.putV8(ctx, c, map[string]interface{}{
		"vms":        PoolName("").mapToString(guestIDs),
		"allow-move": "1"})
}

func (config PoolName) AddGuests(ctx context.Context, c *Client, guestIDs []GuestID) error {
	if err := config.Validate(); err != nil {
		return err
	}
	// TODO: permission check
	exists, err := config.ExistsNoCheck(ctx, c)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New(PoolName_Error_NotExists)
	}
	return config.AddGuestsNoCheck(ctx, c, guestIDs)
}

func (pool PoolName) AddGuestsNoCheck(ctx context.Context, c *Client, guestIDs []GuestID) error {
	version, err := c.GetVersion(ctx)
	if err != nil {
		return err
	}
	config, err := pool.GetNoCheck(ctx, c)
	if err != nil {
		return err
	}
	return pool.addGuestsNoCheck(ctx, c, guestIDs, config.Guests, version)
}

func (config PoolName) Delete(ctx context.Context, c *Client) error {
	if err := config.Validate(); err != nil {
		return err
	}
	// TODO: permission check
	exists, err := config.ExistsNoCheck(ctx, c)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New(PoolName_Error_NotExists)
	}
	return config.DeleteNoCheck(ctx, c)
}

func (config PoolName) DeleteNoCheck(ctx context.Context, c *Client) error {
	if c == nil {
		return errors.New(Client_Error_Nil)
	}
	return c.Delete(ctx, "/pools/"+string(config))
}

func (config PoolName) Exists(ctx context.Context, c *Client) (bool, error) {
	if c == nil {
		return false, errors.New(Client_Error_Nil)
	}
	if err := config.Validate(); err != nil {
		return false, err
	}
	// TODO: permission check
	return config.ExistsNoCheck(ctx, c)
}

func (config PoolName) ExistsNoCheck(ctx context.Context, c *Client) (bool, error) {
	raw, err := listPools(ctx, c)
	if err != nil {
		return false, err
	}
	return ItemInKeyOfArray(raw, "poolid", string(config)), nil
}

func (pool PoolName) Get(ctx context.Context, c *Client) (*ConfigPool, error) {
	if err := pool.Validate(); err != nil {
		return nil, err
	}
	// TODO: permission check
	return pool.GetNoCheck(ctx, c)
}

func (pool PoolName) GetNoCheck(ctx context.Context, c *Client) (*ConfigPool, error) {
	if c == nil {
		return nil, errors.New(Client_Error_Nil)
	}
	params, err := c.GetItemConfigMapStringInterface(ctx, "/pools/"+string(pool), "pool", "CONFIG")
	if err != nil {
		return nil, err
	}
	config := ConfigPool{}.mapToSDK(params)
	return &config, nil
}

func (PoolName) guestsToRemoveFromPools(guests RawGuestResources, guestsToAdd []GuestID) map[PoolName][]GuestID {
	guestsMap := make(map[GuestID]PoolName)
	for i := range guests {
		if v := guests[i].GetPool(); v != "" {
			guestsMap[guests[i].GetID()] = v
		}
	}
	poolMap := make(map[PoolName][]GuestID)
	for i := range guestsToAdd {
		if pool, isSet := guestsMap[guestsToAdd[i]]; isSet {
			if _, isSet := poolMap[pool]; !isSet {
				poolMap[pool] = []GuestID{guestsToAdd[i]}
			} else {
				poolMap[pool] = append(poolMap[pool], guestsToAdd[i])
			}
		}
	}
	return poolMap
}

func (PoolName) mapToString(guestIDs []GuestID) (vms string) {
	for i := range guestIDs {
		vms += "," + strconv.FormatInt(int64(guestIDs[i]), 10)
	}
	if len(vms) > 0 {
		vms = vms[1:]
	}
	return
}

func (pool PoolName) put(ctx context.Context, c *Client, params map[string]interface{}, version Version) error {
	if version.Encode() < version_8_0_0 {
		return pool.putV7(ctx, c, params)
	}
	return pool.putV8(ctx, c, params)
}

func (pool PoolName) putV7(ctx context.Context, c *Client, params map[string]interface{}) error {
	return c.Put(ctx, params, "/pools/"+string(pool))
}

func (pool PoolName) putV8(ctx context.Context, c *Client, params map[string]interface{}) error {
	return c.Put(ctx, params, "/pools?poolid="+string(pool))
}

func (pool PoolName) RemoveGuests(ctx context.Context, c *Client, guestIDs []GuestID) error {
	version, err := c.GetVersion(ctx)
	if err != nil {
		return err
	}
	if err := pool.Validate(); err != nil {
		return err
	}
	if len(guestIDs) == 0 {
		return errors.New(PoolName_Error_NoGuestsSpecified)
	}
	// TODO: permission check
	if exists, err := pool.ExistsNoCheck(ctx, c); err != nil {
		return err
	} else if !exists {
		return errors.New(PoolName_Error_NotExists)
	}
	return pool.removeGuestsNoCheck(ctx, c, guestIDs, version)
}

func (pool PoolName) RemoveGuestsNoChecks(ctx context.Context, c *Client, guestIDs []GuestID) error {
	version, err := c.GetVersion(ctx)
	if err != nil {
		return err
	}
	return pool.removeGuestsNoCheck(ctx, c, guestIDs, version)
}

func (pool PoolName) removeGuestsNoCheck(ctx context.Context, c *Client, guestIDs []GuestID, version Version) error {
	return pool.put(ctx, c, map[string]interface{}{
		"vms":    PoolName("").mapToString(guestIDs),
		"delete": "1"},
		version)
}

func (pool PoolName) SetGuests(ctx context.Context, c *Client, guestIDs []GuestID) error {
	if c == nil {
		return errors.New(Client_Error_Nil)
	}
	if err := pool.Validate(); err != nil {
		return err
	}
	if exists, err := pool.Exists(ctx, c); err != nil {
		return err
	} else if !exists {
		return errors.New(PoolName_Error_NotExists)
	}
	// TODO: permission check
	return pool.SetGuestsNoChecks(ctx, c, guestIDs)
}

func (pool PoolName) SetGuestsNoChecks(ctx context.Context, c *Client, guestID []GuestID) error {
	version, err := c.GetVersion(ctx)
	if err != nil {
		return err
	}
	config, err := pool.Get(ctx, c)
	if err != nil {
		return err
	}
	return pool.setGuestsNoCheck(ctx, c, guestID, config.Guests, version)
}

func (pool PoolName) setGuestsNoCheck(ctx context.Context, c *Client, guestIDs []GuestID, currentGuests *[]GuestID, version Version) error {
	if currentGuests != nil && len(*currentGuests) > 0 {
		if err := pool.removeGuestsNoCheck(ctx, c, subtractArray(*currentGuests, guestIDs), version); err != nil {
			return err
		}
	}
	return pool.addGuestsNoCheck(ctx, c, guestIDs, currentGuests, version)
}

func (pool PoolName) String() string {
	return string(pool)
}

func (config PoolName) Validate() error {
	if config == "" {
		return errors.New(PoolName_Error_Empty)
	}
	if len(config) > 1024 {
		return errors.New(PoolName_Error_Length)
	}
	if !regex_PoolName.MatchString(string(config)) {
		return errors.New(PoolName_Error_Characters)
	}
	return nil
}
