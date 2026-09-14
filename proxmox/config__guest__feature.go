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

func (c *guestClient) HasFeature(ctx context.Context, ref VmRef, feat GuestFeature) (bool, error) {
	if _, err := ref.check_unsafe(ctx, c.api); err != nil {
		return false, err
	}
	return c.HasFeatureNoCheck(ctx, ref, feat)
}

func (c *guestClient) HasFeatureNoCheck(ctx context.Context, ref VmRef, feat GuestFeature) (bool, error) {
	return guestHasFeature(ctx, c.api, ref, feat)
}

func (c *guestClient) ListFeatures(ctx context.Context, ref VmRef) (GuestFeatures, error) {
	if _, err := ref.check_unsafe(ctx, c.api); err != nil {
		return GuestFeatures{}, err
	}
	return c.ListFeaturesNoCheck(ctx, ref)
}

func (c *guestClient) ListFeaturesNoCheck(ctx context.Context, ref VmRef) (feats GuestFeatures, err error) {
	if feats.Clone, err = guestHasFeature(ctx, c.api, ref, GuestFeatureClone); err != nil {
		return
	}
	if feats.Copy, err = guestHasFeature(ctx, c.api, ref, GuestFeatureCopy); err != nil {
		return
	}
	feats.Snapshot, err = guestHasFeature(ctx, c.api, ref, GuestFeatureSnapshot)
	return feats, err
}

func guestHasFeature(ctx context.Context, c *clientAPI, vmr VmRef, feature GuestFeature) (bool, error) {
	params, err := c.getMap(ctx, "/nodes/"+vmr.node.String()+"/"+vmr.vmType.String()+"/"+vmr.vmId.String()+"/feature?feature="+feature.String(), "guest", "FEATURES")
	if err != nil {
		return false, err
	}
	return GuestFeature("").mapToStruct(params), nil
}
