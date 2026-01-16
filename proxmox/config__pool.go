package proxmox

import (
	"context"
	"errors"
	"iter"
	"regexp"
	"strings"

	"github.com/Telmate/proxmox-api-go/internal/array"
	"github.com/Telmate/proxmox-api-go/internal/body"
	"github.com/Telmate/proxmox-api-go/internal/mapUtil"
	"github.com/Telmate/proxmox-api-go/internal/util"
)

type (
	PoolInterface interface {
		AddMembers(context.Context, PoolName, []GuestID, []StorageName) error
		AddMembersNoCheck(context.Context, PoolName, []GuestID, []StorageName) error

		Create(context.Context, ConfigPool) error
		CreateNoCheck(context.Context, ConfigPool) error

		// Delete deletes the specified pool.
		// Returns true if the pool was deleted, false if the pool did not exist.
		Delete(context.Context, PoolName) (bool, error)
		DeleteNoCheck(context.Context, PoolName) (bool, error)

		Exists(context.Context, PoolName) (bool, error)
		ExistsNoCheck(context.Context, PoolName) (bool, error)

		List(ctx context.Context) (RawPools, error)
		ListNoCheck(ctx context.Context) (RawPools, error)

		// When we read the pool config, we also receive a lot of member info.
		Read(context.Context, PoolName) (RawPoolInfo, error)
		ReadNoCheck(context.Context, PoolName) (RawPoolInfo, error)

		RemoveMembers(context.Context, PoolName, []GuestID, []StorageName) error
		RemoveMembersNoCheck(context.Context, PoolName, []GuestID, []StorageName) error

		Set(context.Context, ConfigPool) error
		SetNoCheck(context.Context, ConfigPool) error

		Update(context.Context, ConfigPool) error
		UpdateNoCheck(context.Context, ConfigPool) error
	}

	poolClient struct {
		api       *clientAPI
		oldClient *Client
	}
)

var _ PoolInterface = (*poolClient)(nil)

func (c *poolClient) AddMembers(ctx context.Context, pool PoolName, guests []GuestID, storages []StorageName) error {
	if err := pool.Validate(); err != nil {
		return err
	}
	for i := range guests {
		if err := guests[i].Validate(); err != nil {
			return err
		}
	}
	// TODO check permissions
	return c.AddMembersNoCheck(ctx, pool, guests, storages)
}

func (c *poolClient) AddMembersNoCheck(ctx context.Context, pool PoolName, guests []GuestID, storages []StorageName) error {
	if len(guests) == 0 && len(storages) == 0 {
		return nil
	}
	return pool.addGuests(ctx, c.api, c.oldClient, &guests, &storages, Version{})
}

func (c *poolClient) Create(ctx context.Context, config ConfigPool) error {
	if err := config.Validate(); err != nil {
		return err
	}
	// TODO check permissions
	return c.CreateNoCheck(ctx, config)
}

func (c *poolClient) CreateNoCheck(ctx context.Context, config ConfigPool) error {
	return config.create(ctx, c.api, c.oldClient)
}

func (c *poolClient) Delete(ctx context.Context, pool PoolName) (bool, error) {
	if err := pool.Validate(); err != nil {
		return false, err
	}
	return c.DeleteNoCheck(ctx, pool)
}

func (c *poolClient) DeleteNoCheck(ctx context.Context, pool PoolName) (bool, error) {
	return pool.delete(ctx, c.api)
}

func (c *poolClient) Exists(ctx context.Context, pool PoolName) (bool, error) {
	if err := pool.Validate(); err != nil {
		return false, err
	}
	// TODO check permissions
	return c.ExistsNoCheck(ctx, pool)
}

func (c *poolClient) ExistsNoCheck(ctx context.Context, pool PoolName) (bool, error) {
	return pool.exists(ctx, c.api)
}

func (c *poolClient) List(ctx context.Context) (RawPools, error) {
	// TODO check permissions
	return c.ListNoCheck(ctx)
}

func (c *poolClient) ListNoCheck(ctx context.Context) (RawPools, error) {
	return poolsList(ctx, c.api)
}

func (c *poolClient) Read(ctx context.Context, pool PoolName) (RawPoolInfo, error) {
	if err := pool.Validate(); err != nil {
		return nil, err
	}
	// TODO check permissions
	return c.ReadNoCheck(ctx, pool)
}

func (c *poolClient) ReadNoCheck(ctx context.Context, pool PoolName) (RawPoolInfo, error) {
	raw, errExists, err := pool.read(ctx, c.api)
	if err != nil {
		return nil, err
	}
	if errExists != nil {
		return nil, errExists
	}
	return raw, nil
}

func (c *poolClient) RemoveMembers(ctx context.Context, pool PoolName, guests []GuestID, storages []StorageName) error {
	if err := pool.Validate(); err != nil {
		return err
	}
	for i := range guests {
		if err := guests[i].Validate(); err != nil {
			return err
		}
	}
	// TODO check permissions
	return c.RemoveMembersNoCheck(ctx, pool, guests, storages)
}

func (c *poolClient) RemoveMembersNoCheck(ctx context.Context, pool PoolName, guests []GuestID, storages []StorageName) error {
	return pool.removeMembers(ctx, c.api, &guests, &storages)
}

func (c *poolClient) Set(ctx context.Context, config ConfigPool) error {
	if err := config.Validate(); err != nil {
		return err
	}
	// TODO check permissions
	return c.SetNoCheck(ctx, config)
}

func (c *poolClient) SetNoCheck(ctx context.Context, config ConfigPool) error {
	raw, errExists, err := config.Name.read(ctx, c.api)
	if err != nil {
		return err
	}
	if errExists != nil { // Create
		return config.create(ctx, c.api, c.oldClient)
	}
	return config.update(ctx, c.api, c.oldClient, raw)
}

func (c *poolClient) Update(ctx context.Context, config ConfigPool) error {
	if err := config.Validate(); err != nil {
		return err
	}
	return c.UpdateNoCheck(ctx, config)
}

func (c *poolClient) UpdateNoCheck(ctx context.Context, config ConfigPool) error {
	var current *rawPoolInfo
	if config.Guests != nil || config.Storages != nil {
		var err error
		var errExists error
		current, errExists, err = config.Name.read(ctx, c.api)
		if err != nil {
			return err
		}
		if errExists != nil {
			return errExists
		}
	}
	return config.update(ctx, c.api, c.oldClient, current)
}

// Deprecated: use PoolInterface.List() instead.
func ListPools(ctx context.Context, c *Client) ([]PoolName, error) {
	raw, err := c.New().Pool.List(ctx)
	if err != nil {
		return nil, err
	}
	pools := raw.AsArray()
	poolNames := make([]PoolName, len(pools))
	for i := range pools {
		poolNames[i] = pools[i].GetName()
	}
	return poolNames, nil
}

// Deprecated: use PoolInterface.List() instead.
func ListPoolsWithComments(ctx context.Context, c *Client) (map[PoolName]string, error) {
	raw, err := c.New().Pool.List(ctx)
	if err != nil {
		return nil, err
	}
	pools := raw.AsArray()
	poolMap := make(map[PoolName]string, len(pools))
	for i := range pools {
		poolMap[pools[i].GetName()] = pools[i].GetComment()
	}
	return poolMap, nil
}

type ConfigPool struct {
	Name    PoolName `json:"name"`
	Comment *string  `json:"comment,omitempty"`

	// A guest can only be part of one pool at a time.
	Guests *[]GuestID `json:"guests,omitempty"`

	// A storage can be part of multiple pools at a time.
	Storages *[]StorageName `json:"storages,omitempty"`
}

func (config ConfigPool) mapToApiCreate() *[]byte {
	if config.Comment != nil && *config.Comment != "" {
		return util.Pointer([]byte(poolApiKeyName + "=" + string(config.Name) + "&" + poolApiKeyComment + "=" + body.Escape(*config.Comment)))
	}
	return util.Pointer([]byte(poolApiKeyName + "=" + string(config.Name)))
}

type poolUpdateState int

const (
	poolUpdateStateNoChange poolUpdateState = iota
	poolUpdateStateAdded
	poolUpdateStateRemoved
)

func (config ConfigPool) mapToApiUpdate(current *rawPoolInfo, currentGuests map[GuestID]struct{}, currentStorages map[StorageName]struct{}, version Version) (b *[]byte, guestsAddPtr *[]GuestID, storagesAddPtr *[]StorageName) {
	if current != nil {
		builder := strings.Builder{}
		if config.Comment != nil && *config.Comment != current.GetComment() {
			builder.WriteString("&" + poolApiKeyComment + "=")
			if *config.Comment != "" {
				builder.WriteString(body.Escape(*config.Comment))
			}
		}
		if config.Guests != nil || config.Storages != nil {
			var guestsAdd []GuestID
			var storagesAdd []StorageName
			var guestsRemove []GuestID
			var storagesRemove []StorageName
			guestsAddPtr = &guestsAdd
			storagesAddPtr = &storagesAdd
			if config.Guests != nil {
				guestsAdd, guestsRemove = mapUtil.Difference(array.Map(*config.Guests), currentGuests)
			}
			if config.Storages != nil {
				storagesAdd, storagesRemove = mapUtil.Difference(array.Map(*config.Storages), currentStorages)
			}
			if len(guestsRemove)+len(storagesRemove) > 0 {
				// state = poolUpdateStateRemoved
				builder.WriteString("&delete=1")
				if len(guestsRemove) > 0 {
					builder.WriteString("&" + poolApiKeyGuests + "=")
					builder.WriteString(array.CSV(guestsRemove))
				}
				if len(storagesRemove) > 0 {
					builder.WriteString("&" + poolApiKeyStorages + "=")
					builder.WriteString(array.CSV(storagesRemove))
				}
			} else if len(guestsAdd)+len(storagesAdd) > 0 {
				if len(guestsAdd) > 0 && version.Major >= 8 { // We don't know if the guests are members of another pool
					// state = poolUpdateStateAdded
					builder.WriteString("&allow-move=1&" + poolApiKeyGuests + "=")
					builder.WriteString(array.CSV(guestsAdd))
					guestsAddPtr = nil // we already added them
				}
				if len(storagesAdd) > 0 {
					// state = poolUpdateStateAdded
					builder.WriteString("&" + poolApiKeyStorages + "=")
					builder.WriteString(array.CSV(storagesAdd))
					storagesAddPtr = nil // we already added them
				}
			}
		}
		if builder.Len() == 0 {
			return nil, guestsAddPtr, storagesAddPtr
		}
		return util.Pointer([]byte(builder.String()[1:])), guestsAddPtr, storagesAddPtr
	}
	if config.Comment != nil {
		if *config.Comment == "" {
			return util.Pointer([]byte(poolApiKeyComment + "=")), nil, nil
		} else {
			return util.Pointer([]byte(poolApiKeyComment + "=" + body.Escape(*config.Comment))), nil, nil
		}
	}
	return nil, nil, nil
}

func (config ConfigPool) mapToApi(current *ConfigPool) map[string]any {
	params := map[string]any{}
	if current == nil { //create
		params[poolApiKeyName] = string(config.Name)
		if config.Comment != nil && *config.Comment != "" {
			params[poolApiKeyComment] = string(*config.Comment)
		}
		return params
	}
	// update
	if config.Comment != nil && *config.Comment != *current.Comment {
		params[poolApiKeyComment] = string(*config.Comment)
	}
	return params
}

func (config ConfigPool) create(ctx context.Context, c *clientAPI, oldClient *Client) error {
	if err := c.postRawRetry(ctx, "/pools", config.mapToApiCreate(), 3); err != nil {
		return err
	}
	if config.Guests != nil || config.Storages != nil {
		return config.Name.addGuests(ctx, c, oldClient, config.Guests, config.Storages, Version{})
	}
	return nil
}

// Deprecated: use PoolInterface.Create() instead.
func (config ConfigPool) Create(ctx context.Context, c *Client) error {
	return c.New().Pool.Create(ctx, config)
}

// Deprecated: use PoolInterface.CreateNoCheck() instead.
// CreateNoCheck creates a new pool without validating the input
func (config ConfigPool) CreateNoCheck(ctx context.Context, c *Client) error {
	return c.New().Pool.CreateNoCheck(ctx, config)
}

// Deprecated: use PoolInterface.Delete() instead.
// Same as PoolName.Delete()
func (config ConfigPool) Delete(ctx context.Context, c *Client) error {
	_, err := c.New().Pool.Delete(ctx, config.Name)
	return err
}

// Same as PoolName.Exists()
func (config ConfigPool) Exists(ctx context.Context, c *Client) (bool, error) {
	return config.Name.Exists(ctx, c)
}

// Deprecated: use PoolInterface.Set() instead.
func (config ConfigPool) Set(ctx context.Context, c *Client) error {
	return c.New().Pool.Set(ctx, config)
}

// Deprecated: use PoolInterface.SetNoCheck() instead.
func (config ConfigPool) SetNoCheck(ctx context.Context, c *Client) error {
	return c.New().Pool.SetNoCheck(ctx, config)
}

func (config ConfigPool) update(ctx context.Context, c *clientAPI, oldClient *Client, current *rawPoolInfo) error {
	var version Version
	var guestsMap map[GuestID]struct{}
	var storagesMap map[StorageName]struct{}
	var err error
	if current != nil {
		guestsMap, storagesMap = current.getMembers().maps()
		if config.Guests != nil {
			if mapUtil.SameKeys(guestsMap, array.Map(*config.Guests)) {
				config.Guests = nil
			} else {
				version, err = oldClient.Version(ctx)
				if err != nil {
					return err
				}
			}
		}
	}
	b, guestsAdd, storagesAdd := config.mapToApiUpdate(current, guestsMap, storagesMap, version)
	if b != nil {
		if err = config.Name.put(ctx, c, b); err != nil {
			return err
		}
	}
	return config.Name.addGuests(ctx, c, oldClient, guestsAdd, storagesAdd, version)
}

// Deprecated: use PoolInterface.Update() instead.
func (config ConfigPool) Update(ctx context.Context, c *Client) error {
	return c.New().Pool.Update(ctx, config)
}

// Deprecated: use PoolInterface.UpdateNoCheck() instead.
// UpdateNoCheck updates a pool without validating the input
func (config ConfigPool) UpdateNoCheck(ctx context.Context, c *Client) error {
	return c.New().Pool.UpdateNoCheck(ctx, config)
}

func (config ConfigPool) Validate() error {
	// TODO: Add validation for Guests and Comment
	return config.Name.Validate()
}

type (
	RawConfigPool interface {
		Get() (PoolName, string)
		GetComment() string
		GetName() PoolName
	}
	rawConfigPool struct {
		a    map[string]any
		pool *PoolName
	}
)

var _ RawConfigPool = (*rawConfigPool)(nil)

func (raw *rawConfigPool) Get() (PoolName, string) { return raw.GetName(), raw.GetComment() }

func (raw *rawConfigPool) GetComment() string {
	if v, isSet := raw.a[poolApiKeyComment]; isSet {
		return v.(string)
	}
	return ""
}

func (raw *rawConfigPool) GetName() PoolName {
	if raw.pool != nil {
		return *raw.pool
	}
	if v, isSet := raw.a[poolApiKeyName]; isSet {
		return PoolName(v.(string))
	}
	return ""
}

type (
	RawPools interface {
		AsArray() []RawConfigPool
		AsMap() map[PoolName]RawConfigPool
		Iter() iter.Seq[RawConfigPool]
		Len() int
	}
	rawPools struct{ a []any }
)

var _ RawPools = (*rawPools)(nil)

func (raw *rawPools) AsArray() []RawConfigPool {
	pools := make([]RawConfigPool, len(raw.a))
	for i := range raw.a {
		pools[i] = &rawConfigPool{a: raw.a[i].(map[string]any)}
	}
	return pools
}

func (raw *rawPools) AsMap() map[PoolName]RawConfigPool {
	poolMap := make(map[PoolName]RawConfigPool, len(raw.a))
	for i := range raw.a {
		raw := rawConfigPool{a: raw.a[i].(map[string]any)}
		pool := PoolName(raw.a[poolApiKeyName].(string))
		raw.pool = &pool
		poolMap[pool] = &raw
	}
	return poolMap
}

func (raw *rawPools) Iter() iter.Seq[RawConfigPool] {
	return func(yield func(RawConfigPool) bool) {
		for i := range raw.a {
			if !yield(&rawConfigPool{
				a: raw.a[i].(map[string]any),
			}) {
				return
			}
		}
	}
}

func (raw *rawPools) Len() int { return len(raw.a) }

type (
	RawPoolInfo interface {
		Get() PoolInfo
		GetName() PoolName
		GetComment() string
		GetMembers() RawPoolMembers
	}
	rawPoolInfo struct {
		a map[string]any
	}
)

var _ RawPoolInfo = (*rawPoolInfo)(nil)

func (raw *rawPoolInfo) Get() PoolInfo {
	return PoolInfo{
		Comment: raw.GetComment(),
		Members: raw.GetMembers(),
		Name:    raw.GetName()}
}

func (raw *rawPoolInfo) GetComment() string {
	if v, isSet := raw.a[poolApiKeyComment]; isSet {
		return v.(string)
	}
	return ""
}

func (raw *rawPoolInfo) GetMembers() RawPoolMembers { return raw.getMembers() }

func (raw *rawPoolInfo) getMembers() *rawPoolMembers {
	if v, isSet := raw.a[poolApiKeyMembers]; isSet {
		return &rawPoolMembers{a: v.([]any)}
	}
	return &rawPoolMembers{a: []any{}}
}

func (raw *rawPoolInfo) GetName() PoolName {
	if v, isSet := raw.a[poolApiKeyName]; isSet {
		return PoolName(v.(string))
	}
	return ""
}

type PoolInfo struct {
	Name    PoolName
	Comment string
	Members RawPoolMembers
}

type (
	RawPoolMembers interface {
		AsArray() []RawPoolMember
		AsArrays() ([]RawPoolGuest, []RawPoolStorage)
		AsMaps() (map[GuestID]RawPoolGuest, map[StorageName]RawPoolStorage)
		Iter() iter.Seq[RawPoolMember]
		Len() int
	}
	rawPoolMembers struct {
		a []any
	}
)

var _ RawPoolMembers = (*rawPoolMembers)(nil)

func (raw *rawPoolMembers) AsArray() []RawPoolMember {
	members := make([]RawPoolMember, len(raw.a))
	for i := range raw.a {
		if raw.a[i].(map[string]any)[poolApiKeyMemberType].(string) == "storage" {
			members[i] = &rawPoolMember{a: raw.a[i].(map[string]any)}
		} else {
			members[i] = &rawPoolMember{a: raw.a[i].(map[string]any)}
		}
	}
	return members
}

func (raw *rawPoolMembers) AsArrays() ([]RawPoolGuest, []RawPoolStorage) {
	var storagesCount int
	for i := range raw.a {
		if raw.a[i].(map[string]any)[poolApiKeyMemberType].(string) == "storage" {
			storagesCount++
		}
	}
	// We avoid appending to avoid unnecessary allocations. This would become a problem with very large pools.
	guests := make([]RawPoolGuest, len(raw.a)-storagesCount)
	storages := make([]RawPoolStorage, storagesCount)
	var guestIndex, storageIndex int
	for i := range raw.a {
		params := raw.a[i].(map[string]any)
		if params[poolApiKeyMemberType].(string) == "storage" {
			storages[storageIndex] = &rawPoolStorage{a: params}
			storageIndex++
			continue
		}
		guests[guestIndex] = &rawPoolGuest{a: params}
		guestIndex++
	}
	return guests, storages
}

func (raw *rawPoolMembers) AsMaps() (map[GuestID]RawPoolGuest, map[StorageName]RawPoolStorage) {
	guests := make(map[GuestID]RawPoolGuest)
	storages := make(map[StorageName]RawPoolStorage)
	for i := range raw.a {
		params := raw.a[i].(map[string]any)
		if params[poolApiKeyMemberType].(string) == "storage" {
			name := StorageName(params[poolApiKeyMemberStorageName].(string))
			storages[name] = &rawPoolStorage{a: params, name: &name}
			continue
		}
		guestID := GuestID(int(params[poolApiKeyMemberGuestID].(float64)))
		guests[guestID] = &rawPoolGuest{a: params, id: &guestID}
	}
	return guests, storages
}

func (raw *rawPoolMembers) Iter() iter.Seq[RawPoolMember] {
	return func(yield func(RawPoolMember) bool) {
		for i := range raw.a {
			if !yield(&rawPoolMember{
				a: raw.a[i].(map[string]any),
			}) {
				return
			}
		}
	}
}

func (raw *rawPoolMembers) Len() int { return len(raw.a) }

// Returned pointers are never nil
func (raw *rawPoolMembers) arrays() (*[]GuestID, *[]StorageName) {
	var storagesCount int
	for i := range raw.a {
		if raw.a[i].(map[string]any)[poolApiKeyMemberType].(string) == "storage" {
			storagesCount++
		}
	}
	// We avoid appending to avoid unnecessary allocations. This would become a problem with very large pools.
	guests := make([]GuestID, len(raw.a)-storagesCount)
	storages := make([]StorageName, storagesCount)
	var guestIndex, storageIndex int
	for i := range raw.a {
		params := raw.a[i].(map[string]any)
		if params[poolApiKeyMemberType].(string) == "storage" {
			storages[storageIndex] = StorageName(params[poolApiKeyMemberStorageName].(string))
			storageIndex++
			continue
		}
		guests[guestIndex] = GuestID(int(params[poolApiKeyMemberGuestID].(float64)))
		guestIndex++
	}
	return &guests, &storages
}

func (raw *rawPoolMembers) maps() (map[GuestID]struct{}, map[StorageName]struct{}) {
	guests := make(map[GuestID]struct{})
	storages := make(map[StorageName]struct{})
	for i := range raw.a {
		params := raw.a[i].(map[string]any)
		if params[poolApiKeyMemberType].(string) == "storage" {
			name := StorageName(params[poolApiKeyMemberStorageName].(string))
			storages[name] = struct{}{}
			continue
		}
		guestID := GuestID(int(params[poolApiKeyMemberGuestID].(float64)))
		guests[guestID] = struct{}{}
	}
	return guests, storages
}

type (
	RawPoolMember interface {
		Type() RawPoolMemberType
		AsGuest() (RawPoolGuest, bool)
		AsStorage() (RawPoolStorage, bool)
	}
	rawPoolMember struct {
		a map[string]any
	}
)

var _ RawPoolMember = (*rawPoolMember)(nil)

func (raw *rawPoolMember) Type() RawPoolMemberType {
	if raw.a[poolApiKeyMemberType].(string) == "storage" {
		return RawPoolMemberTypeStorage
	}
	return RawPoolMemberTypeGuest
}

func (raw *rawPoolMember) AsGuest() (RawPoolGuest, bool) {
	if raw.a[poolApiKeyMemberType].(string) == "storage" {
		return nil, false
	}
	return &rawPoolGuest{a: raw.a}, true
}

func (raw *rawPoolMember) AsStorage() (RawPoolStorage, bool) {
	if raw.a[poolApiKeyMemberType].(string) != "storage" {
		return nil, false
	}
	return &rawPoolStorage{a: raw.a}, true
}

// Enum
type RawPoolMemberType int8

const (
	RawPoolMemberTypeUnknown RawPoolMemberType = iota
	RawPoolMemberTypeGuest
	RawPoolMemberTypeStorage
)

type (
	RawPoolGuest interface {
		// TODO parse all information provided about the guest
		GetID() GuestID
	}
	rawPoolGuest struct {
		a  map[string]any
		id *GuestID
	}
)

var _ RawPoolGuest = (*rawPoolGuest)(nil)

func (raw *rawPoolGuest) GetID() GuestID {
	if raw.id != nil {
		return *raw.id
	}
	if v, isSet := raw.a[poolApiKeyMemberGuestID]; isSet {
		return GuestID(v.(float64))
	}
	return 0
}

type (
	RawPoolStorage interface {
		// TODO parse all information provided about the storage
		GetName() StorageName
	}
	rawPoolStorage struct {
		a    map[string]any
		name *StorageName
	}
)

var _ RawPoolStorage = (*rawPoolStorage)(nil)

func (raw *rawPoolStorage) GetName() StorageName {
	if raw.name != nil {
		return *raw.name
	}
	if v, isSet := raw.a[poolApiKeyMemberStorageName]; isSet {
		return StorageName(v.(string))
	}
	return ""
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

func (pool PoolName) addGuests(
	ctx context.Context, c *clientAPI, oldClient *Client,
	guests *[]GuestID, storages *[]StorageName,
	version Version,
) error {
	if (guests == nil || len(*guests) == 0) && (storages == nil || len(*storages) == 0) {
		return nil
	}
	// Only guests got an optimized version in 8.0.0
	// When we only add storages, we can still use the v7 method
	if guests == nil || len(*guests) == 0 {
		return pool.addGuestsV7(ctx, c, nil, storages)
	}
	if version.Major == 0 {
		var err error
		version, err = oldClient.Version(ctx)
		if err != nil {
			return err
		}
	}
	if version.Major >= 8 {
		return pool.addGuestsV8(ctx, c, guests, storages)
	}
	raw, err := listGuests_Unsafe(ctx, c)
	if err != nil {
		return err
	}
	add, remove := guestsToAddAndRemoveFromPools(raw, *guests, pool)
	for i, e := range remove {
		if err = i.removeMembers(ctx, c, &e, nil); err != nil {
			return err
		}
	}
	return pool.addGuestsV7(ctx, c, &add, storages)
}

func (pool PoolName) addGuestsV7(ctx context.Context, c *clientAPI, guests *[]GuestID, storages *[]StorageName) error {
	return pool.addGuestsV(ctx, c, guests, storages, "")
}

// from 8.0.0 on proxmox can move the guests to the pool while they are still in another pool
func (pool PoolName) addGuestsV8(ctx context.Context, c *clientAPI, guests *[]GuestID, storages *[]StorageName) error {
	return pool.addGuestsV(ctx, c, guests, storages, "allow-move=1&")
}

func (pool PoolName) addGuestsV(ctx context.Context, c *clientAPI, guests *[]GuestID, storages *[]StorageName, move string) error {
	builder := strings.Builder{}
	if guests != nil && len(*guests) > 0 {
		builder.WriteString(move + poolApiKeyGuests + "=")
		builder.WriteString(array.CSV(*guests))
	}
	if storages != nil && len(*storages) > 0 {
		if guests != nil && len(*guests) > 0 {
			builder.WriteString("&")
		}
		builder.WriteString(poolApiKeyStorages + "=")
		builder.WriteString(array.CSV(*storages))
	}
	return pool.put(ctx, c, util.Pointer([]byte(builder.String())))
}

// Deprecated: use PoolInterface.AddGuests() instead.
func (pool PoolName) AddGuests(ctx context.Context, c *Client, guestIDs []GuestID) error {
	return c.New().Pool.AddMembers(ctx, pool, guestIDs, nil)
}

// Deprecated: use PoolInterface.AddGuestsNoCheck() instead.
func (pool PoolName) AddGuestsNoCheck(ctx context.Context, c *Client, guestIDs []GuestID) error {
	return c.New().Pool.AddMembersNoCheck(ctx, pool, guestIDs, nil)
}

func (pool PoolName) delete(ctx context.Context, c *clientAPI) (bool, error) {
	if err := c.deleteRetry(ctx, "/pools/"+pool.String(), 3); err != nil {
		var apiErr *ApiError
		if errors.As(err, &apiErr) {
			const prefix = "delete pool failed: pool '"
			const prefixLen = len(prefix)
			if strings.HasPrefix(apiErr.Message, prefix) {
				if strings.HasPrefix(apiErr.Message[prefixLen:], pool.String()+"' is not empty") {
					if err = pool.empty(ctx, c); err != nil {
						return false, err
					}
					if err = c.deleteRetry(ctx, "/pools/"+pool.String(), 3); err != nil {
						return false, err
					}
					return true, nil
				}
				if strings.HasPrefix(apiErr.Message[prefixLen:], pool.String()+"' does not exist") {
					return false, nil
				}
			}
		}
		return false, err
	}
	return true, nil
}

// Deprecated: use PoolInterface.Delete() instead.
func (pool PoolName) Delete(ctx context.Context, c *Client) error {
	_, err := c.New().Pool.Delete(ctx, pool)
	return err
}

// Deprecated: use PoolInterface.DeleteNoCheck() instead.
func (pool PoolName) DeleteNoCheck(ctx context.Context, c *Client) error {
	_, err := c.New().Pool.DeleteNoCheck(ctx, pool)
	return err
}

func (pool PoolName) empty(ctx context.Context, c *clientAPI) error {
	raw, errExists, err := pool.read(ctx, c)
	if err != nil {
		return err
	}
	if errExists != nil {
		return errExists
	}
	guests, storages := raw.getMembers().arrays()
	return pool.removeMembers(ctx, c, guests, storages)
}

func (pool PoolName) exists(ctx context.Context, c *clientAPI) (bool, error) {
	_, exists, err := pool.read(ctx, c)
	if err != nil {
		return false, err
	}
	return exists == nil, nil
}

// Deprecated: use PoolInterface.Exists() instead.
func (pool PoolName) Exists(ctx context.Context, c *Client) (bool, error) {
	return c.New().Pool.Exists(ctx, pool)
}

// Deprecated: use PoolInterface.ExistsNoCheck() instead.
func (pool PoolName) ExistsNoCheck(ctx context.Context, c *Client) (bool, error) {
	return c.New().Pool.ExistsNoCheck(ctx, pool)
}

// Deprecated: use PoolInterface.Read() instead.
func (pool PoolName) Get(ctx context.Context, c *Client) (RawPoolInfo, error) {
	return c.New().Pool.Read(ctx, pool)
}

// Deprecated: use PoolInterface.ReadNoCheck() instead.
func (pool PoolName) GetNoCheck(ctx context.Context, c *Client) (RawPoolInfo, error) {
	return c.New().Pool.ReadNoCheck(ctx, pool)
}

func (pool PoolName) put(ctx context.Context, c *clientAPI, body *[]byte) error {
	return c.putRawRetry(ctx, "/pools/"+pool.String(), body, 3)
}

// errExists is non-nil if the pool does not exist.
func (pool PoolName) read(ctx context.Context, c *clientAPI) (r *rawPoolInfo, errExists error, err error) {
	var raw map[string]any
	raw, err = c.getMap(ctx, "/pools/"+pool.String(), "pool", "CONFIG")
	if err != nil {
		var apiErr *ApiError
		if errors.As(err, &apiErr) { // check for not found
			if strings.HasPrefix(apiErr.Message, "pool '"+pool.String()+"' does not exist") {
				return nil, err, nil
			}
		}
		return nil, nil, err
	}
	return &rawPoolInfo{a: raw}, nil, nil
}

func (pool PoolName) removeMembers(ctx context.Context, c *clientAPI, guests *[]GuestID, storages *[]StorageName) error {
	if (guests == nil || len(*guests) == 0) && (storages == nil || len(*storages) == 0) {
		return nil
	}
	builder := strings.Builder{}
	if guests != nil && len(*guests) > 0 {
		builder.WriteString(poolApiKeyGuests + "=")
		builder.WriteString(array.CSV(*guests))
	}
	if storages != nil && len(*storages) > 0 {
		if guests != nil && len(*guests) > 0 {
			builder.WriteString("&")
		}
		builder.WriteString(poolApiKeyStorages + "=")
		builder.WriteString(array.CSV(*storages))
	}
	builder.WriteString("&delete=1")
	return pool.put(ctx, c, util.Pointer([]byte(builder.String())))
}

// Deprecated: use PoolInterface.RemoveGuests() instead.
func (pool PoolName) RemoveGuests(ctx context.Context, c *Client, guestIDs []GuestID) error {
	return c.New().Pool.RemoveMembers(ctx, pool, guestIDs, nil)
}

// Deprecated: use PoolInterface.RemoveGuestsNoCheck() instead.
func (pool PoolName) RemoveGuestsNoChecks(ctx context.Context, c *Client, guestIDs []GuestID) error {
	return c.New().Pool.RemoveMembersNoCheck(ctx, pool, guestIDs, nil)
}

// Deprecated: use PoolInterface.Set() instead.
func (pool PoolName) SetGuests(ctx context.Context, c *Client, guestIDs []GuestID) error {
	return c.New().Pool.Set(ctx, ConfigPool{Name: pool, Guests: &guestIDs})
}

// Deprecated: use PoolInterface.SetNoCheck() instead.
func (pool PoolName) SetGuestsNoChecks(ctx context.Context, c *Client, guestID []GuestID) error {
	return c.New().Pool.SetNoCheck(ctx, ConfigPool{Name: pool, Guests: &guestID})
}

func (pool PoolName) String() string { return string(pool) } // String is for fmt.Stringer.

func (pool PoolName) Validate() error {
	if pool == "" {
		return errors.New(PoolName_Error_Empty)
	}
	if len(pool) > 1024 {
		return errors.New(PoolName_Error_Length)
	}
	if !regex_PoolName.MatchString(string(pool)) {
		return errors.New(PoolName_Error_Characters)
	}
	return nil
}

func guestsToAddAndRemoveFromPools(guests RawGuestResources, guestsToAdd []GuestID, targetPool PoolName) ([]GuestID, map[PoolName][]GuestID) {
	guestsMap := make(map[GuestID]RawGuestResource)
	for i := range guests {
		guestsMap[guests[i].GetID()] = guests[i]
	}
	add := make([]GuestID, 0, len(guestsToAdd))
	remove := make(map[PoolName][]GuestID)
	for _, id := range guestsToAdd {
		guest, exists := guestsMap[id]
		if !exists { // Case 1: guest does not exist
			add = append(add, id)
			continue
		}
		pool := guest.GetPool()
		if pool == targetPool { // Case 2: already in target pool
			continue
		}
		// Case 3: exists, different pool
		add = append(add, id)
		if pool != "" {
			remove[pool] = append(remove[pool], id)
		}
	}
	return add, remove
}

func poolsList(ctx context.Context, c *clientAPI) (*rawPools, error) {
	raw, err := c.getList(ctx, "/pools", "pools", "list")
	if err != nil {
		return nil, err
	}
	return &rawPools{a: raw}, nil
}

const (
	poolApiKeyMembers string = "members"
	poolApiKeyComment string = "comment"
	poolApiKeyName    string = "poolid"

	poolApiKeyMemberType string = "type"

	poolApiKeyMemberStorageName string = "storage"
	poolApiKeyMemberGuestID     string = "vmid"

	poolApiKeyGuests   string = "vms"
	poolApiKeyStorages string = "storage"
)
