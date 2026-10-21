package proxmox

import (
	"context"
	"errors"
)

// GuestFeature is an enum
//
//	const (
//		GuestFeatureClone
//		GuestFeatureCopy
//		GuestFeatureSnapshot
//	)
type GuestFeature string

const (
	GuestFeatureClone    GuestFeature = "clone"
	GuestFeatureCopy     GuestFeature = "copy"
	GuestFeatureSnapshot GuestFeature = "snapshot"
)

func (GuestFeature) Error() error {
	return errors.New("value should be one of (" + string(GuestFeatureClone) + " ," + string(GuestFeatureCopy) + " ," + string(GuestFeatureSnapshot) + ")")
}

func (GuestFeature) mapToStruct(params map[string]interface{}) bool {
	if value, isSet := params["hasFeature"]; isSet {
		return int(value.(float64)) == 1
	}
	return false
}

func (f GuestFeature) String() string { return string(f) }

func (f GuestFeature) Validate() error {
	switch f {
	case GuestFeatureCopy, GuestFeatureClone, GuestFeatureSnapshot:
		return nil
	}
	return GuestFeature("").Error()
}

type GuestFeatures struct {
	Clone    bool `json:"clone"`
	Copy     bool `json:"copy"`
	Snapshot bool `json:"snapshot"`
}

func (c *guestClient) HasFeature(ctx context.Context, vmr VmRef, feat GuestFeature) (bool, error) {
	ref := vmr.toGuestRef()
	if _, err := ref.check(ctx, nil, c.api); err != nil {
		return false, err
	}
	return c.HasFeatureNoCheck(ctx, ref.legacy(), feat)
}

func (c *guestClient) HasFeatureNoCheck(ctx context.Context, ref VmRef, feat GuestFeature) (bool, error) {
	return ref.toGuestRef().hasFeature(ctx, c.api, feat)
}

func (c *guestClient) ListFeatures(ctx context.Context, vmr VmRef) (GuestFeatures, error) {
	ref := vmr.toGuestRef()
	if _, err := ref.check(ctx, nil, c.api); err != nil {
		return GuestFeatures{}, err
	}
	return c.ListFeaturesNoCheck(ctx, ref.legacy())
}

func (c *guestClient) ListFeaturesNoCheck(ctx context.Context, vmr VmRef) (feats GuestFeatures, err error) {
	ref := vmr.toGuestRef()
	if feats.Clone, err = ref.hasFeature(ctx, c.api, GuestFeatureClone); err != nil {
		return
	}
	if feats.Copy, err = ref.hasFeature(ctx, c.api, GuestFeatureCopy); err != nil {
		return
	}
	feats.Snapshot, err = ref.hasFeature(ctx, c.api, GuestFeatureSnapshot)
	return feats, err
}

func (ref guestRef) hasFeature(ctx context.Context, c *clientAPI, feature GuestFeature) (bool, error) {
	params, err := c.getMap(ctx, "/nodes/"+ref.Node.String()+"/"+ref.Type.String()+"/"+ref.ID.String()+"/feature?feature="+feature.String(), "guest", "FEATURES")
	if err != nil {
		return false, err
	}
	return GuestFeature("").mapToStruct(params), nil
}
