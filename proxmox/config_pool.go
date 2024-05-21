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

func (config ConfigPool) Create(c *Client) error {
	if err := config.Validate(); err != nil {
		return err
	}
	// TODO check permissions
	if exists, err := config.Name.Exists_Unsafe(c); err != nil {
		return err
	} else if exists {
		return errors.New(PoolName_Error_Exists)
	}
	return config.Create_Unsafe(c)
}

// Create_Unsafe creates a new pool without validating the input
func (config ConfigPool) Create_Unsafe(c *Client) error {
	version, err := c.GetVersion()
	if err != nil {
		return err
	}
	if err := c.Post(config.mapToApi(nil), "/pools"); err != nil {
		return err
	}
	if config.Guests != nil {
		return config.Name.addGuests_Unsafe(c, *config.Guests, nil, version)
	}
	return nil
}

// Same as PoolName.Delete()
func (config ConfigPool) Delete(c *Client) error {
	return config.Name.Delete(c)
}

// Same as PoolName.Exists()
func (config ConfigPool) Exists(c *Client) (bool, error) {
	return config.Name.Exists(c)
}

func (config ConfigPool) Set(c *Client) error {
	if c == nil {
		return errors.New(Client_Error_Nil)
	}
	if err := config.Validate(); err != nil {
		return err
	}
	// TODO check permissions
	return config.Set_Unsafe(c)
}

func (config ConfigPool) Set_Unsafe(c *Client) error {
	exists, err := config.Name.Exists_Unsafe(c)
	if err != nil {
		return err
	}
	if exists {
		current, err := config.Name.Get_Unsafe(c)
		if err != nil {
			return err
		}
		return config.Update_Unsafe(c, current)
	}
	return config.Create_Unsafe(c)
}

func (config ConfigPool) Update(c *Client) error {
	if c == nil {
		return errors.New(Client_Error_Nil)
	}
	if err := config.Validate(); err != nil {
		return err
	}
	if exists, err := config.Name.Exists(c); err != nil {
		return err
	} else if !exists {
		return errors.New(PoolName_Error_NotExists)
	}
	// TODO check permissions
	current, err := config.Name.Get_Unsafe(c)
	if err != nil {
		return err
	}
	return config.Update_Unsafe(c, current)
}

// Update_Unsafe updates a pool without validating the input
func (config ConfigPool) Update_Unsafe(c *Client, current *ConfigPool) error {
	if params := config.mapToApi(current); len(params) > 0 {
		if err := config.Name.put(c, params); err != nil {
			return err
		}
	}
	if config.Guests != nil {
		return config.Name.SetGuests_Unsafe(c, *config.Guests)
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

func (config PoolName) addGuests_Unsafe(c *Client, guestIDs []uint, currentGuests *[]uint, version Version) error {
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
		return config.addGuests_UnsafeV8(c, guestsToAdd)
	}
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

func (pool PoolName) addGuests_UnsafeV7(c *Client, guestIDs []uint) error {
	return pool.put(c, map[string]interface{}{"vms": PoolName("").mapToString(guestIDs)})
}

func (pool PoolName) addGuests_UnsafeV8(c *Client, guestIDs []uint) error {
	return pool.put(c, map[string]interface{}{
		"vms":        PoolName("").mapToString(guestIDs),
		"allow-move": "1"})
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
	version, err := c.GetVersion()
	if err != nil {
		return err
	}
	config, err := pool.Get_Unsafe(c)
	if err != nil {
		return err
	}
	return pool.addGuests_Unsafe(c, guestIDs, config.Guests, version)
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

func (pool PoolName) put(c *Client, params map[string]interface{}) error {
	return c.Put(params, "/pools?poolid="+string(pool))
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

func (pool PoolName) RemoveGuests_Unsafe(c *Client, guestID []uint) error {
	return pool.put(c, map[string]interface{}{
		"vms":    PoolName("").mapToString(guestID),
		"delete": "1"})
}

func (pool PoolName) SetGuests(c *Client, guestID []uint) error {
	if c == nil {
		return errors.New(Client_Error_Nil)
	}
	if err := pool.Validate(); err != nil {
		return err
	}
	if exists, err := pool.Exists(c); err != nil {
		return err
	} else if !exists {
		return errors.New(PoolName_Error_NotExists)
	}
	// TODO: permission check
	return pool.SetGuests_Unsafe(c, guestID)
}

func (pool PoolName) SetGuests_Unsafe(c *Client, guestID []uint) error {
	version, err := c.GetVersion()
	if err != nil {
		return err
	}
	config, err := pool.Get(c)
	if err != nil {
		return err
	}
	return pool.setGuests_Unsafe(c, guestID, config.Guests, version)
}

func (pool PoolName) setGuests_Unsafe(c *Client, guestIDs []uint, currentGuests *[]uint, version Version) error {
	var guestsToRemove []uint
	for _, e := range *currentGuests {
		removeGuest := true
		for _, ee := range guestIDs {
			if e == ee {
				removeGuest = false
				break
			}
		}
		if removeGuest {
			guestsToRemove = append(guestsToRemove, e)
		}
	}
	if err := pool.RemoveGuests_Unsafe(c, guestsToRemove); err != nil {
		return err
	}
	return pool.addGuests_Unsafe(c, guestIDs, currentGuests, version)
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
