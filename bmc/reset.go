package bmc

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// BMCResetter for resetting a BMC.
// resetType: "warm" resets the management console without rebooting the BMC
// resetType: "cold" reboots the BMC
type BMCResetter interface {
	BmcReset(ctx context.Context, resetType string) (ok bool, err error)
}

// bmcProviders is an internal struct to correlate an implementation/provider and its name
type bmcProviders struct {
	name        string
	bmcResetter BMCResetter
}

// resetBMC tries all implementations for a success BMC reset
func resetBMC(ctx context.Context, resetType string, b []bmcProviders) (ok bool, metadata Metadata, err error) {
	var metadataLocal Metadata
Loop:
	for _, elem := range b {
		if elem.bmcResetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			ok, setErr := elem.bmcResetter.BmcReset(ctx, resetType)
			if setErr != nil {
				err = multierror.Append(err, errors.WithMessagef(setErr, "provider: %v", elem.name))
				continue
			}
			if !ok {
				err = multierror.Append(err, fmt.Errorf("provider: %v, failed to reset BMC", elem.name))
				continue
			}
			metadataLocal.SuccessfulProvider = elem.name
			return ok, metadataLocal, nil
		}
	}
	return ok, metadataLocal, multierror.Append(err, errors.New("failed to reset BMC"))
}

// ResetBMCFromInterfaces identifies implementations of the BMCResetter interface and passes them to the resetBMC() wrapper method.
func ResetBMCFromInterfaces(ctx context.Context, resetType string, generic []interface{}) (ok bool, metadata Metadata, err error) {
	bmcSetters := make([]bmcProviders, 0)
	for _, elem := range generic {
		temp := bmcProviders{name: getProviderName(elem)}
		switch p := elem.(type) {
		case BMCResetter:
			temp.bmcResetter = p
			bmcSetters = append(bmcSetters, temp)
		default:
			e := fmt.Sprintf("not a BMCResetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(bmcSetters) == 0 {
		return ok, metadata, multierror.Append(err, errors.New("no BMCResetter implementations found"))
	}
	return resetBMC(ctx, resetType, bmcSetters)
}
