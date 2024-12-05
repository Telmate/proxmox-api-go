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
	Name    PoolName `json:"name"`
	Comment *string  `json:"comment"`
	Guests  *[]uint  `json:"guests"` // TODO: Change type once we have a type for guestID
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
		guests := make([]uint, 0)
		for _, e := range v.([]interface{}) {
			param := e.(map[string]interface{})
			if v, isSet := param["vmid"]; isSet {
				guests = append(guests, uint(v.(float64)))
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
	if exists, err := config.Name.Exists_Unsafe(ctx, c); err != nil {
		return err
	} else if exists {
		return errors.New(PoolName_Error_Exists)
	}
	return config.Create_Unsafe(ctx, c)
}

// Create_Unsafe creates a new pool without validating the input
func (config ConfigPool) Create_Unsafe(ctx context.Context, c *Client) error {
	version, err := c.GetVersion(ctx)
	if err != nil {
		return err
	}
	if err := c.Post(ctx, config.mapToApi(nil), "/pools"); err != nil {
		return err
	}
	if config.Guests != nil {
		return config.Name.addGuests_Unsafe(ctx, c, *config.Guests, nil, version)
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
	return config.Set_Unsafe(ctx, c)
}

func (config ConfigPool) Set_Unsafe(ctx context.Context, c *Client) error {
	exists, err := config.Name.Exists_Unsafe(ctx, c)
	if err != nil {
		return err
	}
	if exists {
		return config.Update_Unsafe(ctx, c)
	}
	return config.Create_Unsafe(ctx, c)
}

func (config ConfigPool) Update(ctx context.Context, c *Client) error {
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
	return config.Update_Unsafe(ctx, c)
}

// Update_Unsafe updates a pool without validating the input
func (config ConfigPool) Update_Unsafe(ctx context.Context, c *Client) error {
	version, err := c.GetVersion(ctx)
	if err != nil {
		return err
	}
	current, err := config.Name.Get_Unsafe(ctx, c)
	if err != nil {
		return err
	}
	if params := config.mapToApi(current); len(params) > 0 {
		if err = config.Name.put(ctx, c, params, version); err != nil {
			return err
		}
	}
	if config.Guests != nil {
		return config.Name.SetGuests_Unsafe(ctx, c, *config.Guests)
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

func (config PoolName) addGuests_Unsafe(ctx context.Context, c *Client, guestIDs []uint, currentGuests *[]uint, version Version) error {
	var guestsToAdd []uint
	if currentGuests != nil && len(*currentGuests) > 0 {
		guestsToAdd = subtractArray(guestIDs, *currentGuests)
	} else {
		guestsToAdd = guestIDs
	}
	if len(guestsToAdd) == 0 {
		return nil
	}
	if !version.Smaller(Version{8, 0, 0}) {
		return config.addGuests_UnsafeV8(ctx, c, guestsToAdd)
	}
	guests, err := ListGuests(ctx, c)
	if err != nil {
		return err
	}
	for i, e := range PoolName("").guestsToRemoveFromPools(guests, guestsToAdd) {
		if err = i.removeGuests_Unsafe(ctx, c, e, version); err != nil {
			return err
		}
	}
	return config.addGuests_UnsafeV7(ctx, c, guestsToAdd)
}

func (pool PoolName) addGuests_UnsafeV7(ctx context.Context, c *Client, guestIDs []uint) error {
	return pool.putV7(ctx, c, map[string]interface{}{"vms": PoolName("").mapToString(guestIDs)})
}

// from 8.0.0 on proxmox can move the guests to the pool while they are still in another pool
func (pool PoolName) addGuests_UnsafeV8(ctx context.Context, c *Client, guestIDs []uint) error {
	return pool.putV8(ctx, c, map[string]interface{}{
		"vms":        PoolName("").mapToString(guestIDs),
		"allow-move": "1"})
}

func (config PoolName) AddGuests(ctx context.Context, c *Client, guestIDs []uint) error {
	if err := config.Validate(); err != nil {
		return err
	}
	// TODO: permission check
	exists, err := config.Exists_Unsafe(ctx, c)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New(PoolName_Error_NotExists)
	}
	return config.AddGuests_Unsafe(ctx, c, guestIDs)
}

func (pool PoolName) AddGuests_Unsafe(ctx context.Context, c *Client, guestIDs []uint) error {
	version, err := c.GetVersion(ctx)
	if err != nil {
		return err
	}
	config, err := pool.Get_Unsafe(ctx, c)
	if err != nil {
		return err
	}
	return pool.addGuests_Unsafe(ctx, c, guestIDs, config.Guests, version)
}

func (config PoolName) Delete(ctx context.Context, c *Client) error {
	if err := config.Validate(); err != nil {
		return err
	}
	// TODO: permission check
	exists, err := config.Exists_Unsafe(ctx, c)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New(PoolName_Error_NotExists)
	}
	return config.Delete_Unsafe(ctx, c)
}

func (config PoolName) Delete_Unsafe(ctx context.Context, c *Client) error {
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
	return config.Exists_Unsafe(ctx, c)
}

func (config PoolName) Exists_Unsafe(ctx context.Context, c *Client) (bool, error) {
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
	return pool.Get_Unsafe(ctx, c)
}

func (pool PoolName) Get_Unsafe(ctx context.Context, c *Client) (*ConfigPool, error) {
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

func (PoolName) guestsToRemoveFromPools(guests []GuestResource, guestsToAdd []uint) map[PoolName][]uint {
	// map[guestID]PoolName
	guestsMap := make(map[uint]PoolName)
	for _, e := range guests {
		if e.Pool != "" {
			guestsMap[e.Id] = e.Pool
			continue
		}
	}
	poolMap := make(map[PoolName][]uint)
	for _, e := range guestsToAdd {
		if pool, isSet := guestsMap[e]; isSet {
			if _, isSet := poolMap[pool]; !isSet {
				poolMap[pool] = []uint{e}
			} else {
				poolMap[pool] = append(poolMap[pool], e)
			}
		}
	}
	return poolMap
}

// TODO replace once we have a type for guestID
func (PoolName) mapToString(guestID []uint) (vms string) {
	for _, e := range guestID {
		vms += "," + strconv.FormatInt(int64(e), 10)
	}
	if len(vms) > 0 {
		vms = vms[1:]
	}
	return
}

func (pool PoolName) put(ctx context.Context, c *Client, params map[string]interface{}, version Version) error {
	if version.Smaller(Version{8, 0, 0}) {
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

func (pool PoolName) RemoveGuests(ctx context.Context, c *Client, guestID []uint) error {
	version, err := c.GetVersion(ctx)
	if err != nil {
		return err
	}
	if err := pool.Validate(); err != nil {
		return err
	}
	if len(guestID) == 0 {
		return errors.New(PoolName_Error_NoGuestsSpecified)
	}
	// TODO: permission check
	if exists, err := pool.Exists_Unsafe(ctx, c); err != nil {
		return err
	} else if !exists {
		return errors.New(PoolName_Error_NotExists)
	}
	return pool.removeGuests_Unsafe(ctx, c, guestID, version)
}

func (pool PoolName) RemoveGuests_Unsafe(ctx context.Context, c *Client, guestID []uint) error {
	version, err := c.GetVersion(ctx)
	if err != nil {
		return err
	}
	return pool.removeGuests_Unsafe(ctx, c, guestID, version)
}

func (pool PoolName) removeGuests_Unsafe(ctx context.Context, c *Client, guestID []uint, version Version) error {
	return pool.put(ctx, c, map[string]interface{}{
		"vms":    PoolName("").mapToString(guestID),
		"delete": "1"},
		version)
}

func (pool PoolName) SetGuests(ctx context.Context, c *Client, guestID []uint) error {
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
	return pool.SetGuests_Unsafe(ctx, c, guestID)
}

func (pool PoolName) SetGuests_Unsafe(ctx context.Context, c *Client, guestID []uint) error {
	version, err := c.GetVersion(ctx)
	if err != nil {
		return err
	}
	config, err := pool.Get(ctx, c)
	if err != nil {
		return err
	}
	return pool.setGuests_Unsafe(ctx, c, guestID, config.Guests, version)
}

func (pool PoolName) setGuests_Unsafe(ctx context.Context, c *Client, guestIDs []uint, currentGuests *[]uint, version Version) error {
	if currentGuests != nil && len(*currentGuests) > 0 {
		if err := pool.removeGuests_Unsafe(ctx, c, subtractArray(*currentGuests, guestIDs), version); err != nil {
			return err
		}
	}
	return pool.addGuests_Unsafe(ctx, c, guestIDs, currentGuests, version)
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
