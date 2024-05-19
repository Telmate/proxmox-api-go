package proxmox

import (
	"errors"
	"regexp"
	"strconv"
)

func ListPools(c *Client) ([]PoolName, error) {
	raw, err := listPools(c)
	if err != nil {
		return nil, err
	}
	pools := make([]PoolName, len(raw))
	for i, e := range raw {
		pools[i] = PoolName(e.(map[string]interface{})["poolid"].(string))
	}
	return pools, nil
}

func listPools(c *Client) ([]interface{}, error) {
	return c.GetItemListInterfaceArray("/pools")
}

type ConfigPool struct {
	Name    PoolName `json:"name"`
	Comment *string  `json:"comment"`
	Guests  *[]uint  `json:"guests"` // TODO: Change type once we have a type for guestID
}

func (ConfigPool) mapToSDK(params map[string]interface{}) (config ConfigPool) {
	if v, isSet := params["poolid"]; isSet {
		config.Name = PoolName(v.(string))
	}
	if v, isSet := params["comment"]; isSet {
		tmp := v.(string)
		config.Comment = &tmp
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

// Same as PoolName.Delete()
func (config ConfigPool) Delete(c *Client) error {
	return config.Name.Delete(c)
}

// Same as PoolName.Exists()
func (config ConfigPool) Exists(c *Client) (bool, error) {
	return config.Name.Exists(c)
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
	PoolName_Error_NoGuestsSpecified string = "no guests specified"
)

var regex_PoolName = regexp.MustCompile(`^[a-zA-Z0-9-_]+$`)

func (config PoolName) addGuests_Unsafe(c *Client, guestIDs []uint, currentGuests *[]uint) error {
	var guestsToAdd []uint
	if currentGuests != nil && len(*currentGuests) > 0 {
		guestsToAdd = subtractArray(guestIDs, *currentGuests)
	} else {
		guestsToAdd = guestIDs
	}
	if len(guestsToAdd) == 0 {
		return nil
	}
	if c.version.Smaller(Version{8, 0, 0}) {
		guests, err := ListGuests(c)
		if err != nil {
			return err
		}
		for i, e := range PoolName("").guestsToRemoveFromPools(guests, guestsToAdd) {
			if err = i.RemoveGuests_Unsafe(c, e); err != nil {
				return err
			}
		}
		return config.addGuests_UnsafeV7(c, guestsToAdd)
	}
	return config.addGuests_UnsafeV8(c, guestsToAdd)
}

func (config PoolName) addGuests_UnsafeV7(c *Client, guestIDs []uint) error {
	params := config.mapToApi(guestIDs)
	return c.Put(params, "/pools")
}

func (config PoolName) addGuests_UnsafeV8(c *Client, guestIDs []uint) error {
	params := config.mapToApi(guestIDs)
	params["allow-move"] = "1"
	return c.Put(params, "/pools")
}

func (config PoolName) AddGuests(c *Client, guestIDs []uint) error {
	if c == nil {
		return errors.New(Client_Error_Nil)
	}
	if err := config.Validate(); err != nil {
		return err
	}
	// TODO: permission check
	exists, err := config.Exists_Unsafe(c)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New(PoolName_Error_NotExists)
	}
	return config.AddGuests_Unsafe(c, guestIDs)
}

func (pool PoolName) AddGuests_Unsafe(c *Client, guestIDs []uint) error {
	config, err := pool.Get_Unsafe(c)
	if err != nil {
		return err
	}
	return pool.addGuests_Unsafe(c, guestIDs, config.Guests)
}

func (config PoolName) Delete(c *Client) error {
	if c == nil {
		return errors.New(Client_Error_Nil)
	}
	if err := config.Validate(); err != nil {
		return err
	}
	// TODO: permission check
	exists, err := config.Exists_Unsafe(c)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New(PoolName_Error_NotExists)
	}
	return config.Delete_Unsafe(c)
}

func (config PoolName) Delete_Unsafe(c *Client) error {
	return c.Delete("/pools/" + string(config))
}

func (config PoolName) Exists(c *Client) (bool, error) {
	if c == nil {
		return false, errors.New(Client_Error_Nil)
	}
	if err := config.Validate(); err != nil {
		return false, err
	}
	// TODO: permission check
	return config.Exists_Unsafe(c)
}

func (config PoolName) Exists_Unsafe(c *Client) (bool, error) {
	raw, err := listPools(c)
	if err != nil {
		return false, err
	}
	return ItemInKeyOfArray(raw, "poolid", string(config)), nil
}

func (pool PoolName) Get(c *Client) (*ConfigPool, error) {
	if c == nil {
		return nil, errors.New(Client_Error_Nil)
	}
	if err := pool.Validate(); err != nil {
		return nil, err
	}
	// TODO: permission check
	return pool.Get_Unsafe(c)
}

func (pool PoolName) Get_Unsafe(c *Client) (*ConfigPool, error) {
	params, err := c.GetItemConfigMapStringInterface("/pools/"+string(pool), "pool", "CONFIG")
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

func (config PoolName) mapToApi(guestID []uint) map[string]interface{} {
	var vms string
	for _, e := range guestID {
		vms += "," + strconv.FormatInt(int64(e), 10)
	}
	if len(vms) > 0 {
		vms = vms[1:]
	}
	return map[string]interface{}{
		"poolid": string(config),
		"vms":    vms,
	}
}

func (config PoolName) RemoveGuests(c *Client, guestID []uint) error {
	if c == nil {
		return errors.New(Client_Error_Nil)
	}
	if err := config.Validate(); err != nil {
		return err
	}
	if len(guestID) == 0 {
		return errors.New(PoolName_Error_NoGuestsSpecified)
	}
	// TODO: permission check
	if exists, err := config.Exists_Unsafe(c); err != nil {
		return err
	} else if !exists {
		return errors.New(PoolName_Error_NotExists)
	}
	return config.RemoveGuests_Unsafe(c, guestID)
}

func (config PoolName) RemoveGuests_Unsafe(c *Client, guestID []uint) error {
	params := config.mapToApi(guestID)
	params["delete"] = "1"
	return c.Put(params, "/pools")
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
