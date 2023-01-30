package redfishwrapper

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	rf "github.com/stmcginnis/gofish/redfish"
)

// Set the boot device for the system.
func (c *Client) SetVirtualMedia(ctx context.Context, kind string, mediaUrl string) (ok bool, err error) {
	managers, err := c.Managers(ctx)
	if err != nil {
		return false, err
	}

	var mediaKind rf.VirtualMediaType
	switch kind {
	case "CD":
		mediaKind = rf.CDMediaType
	case "Floppy":
		mediaKind = rf.FloppyMediaType
	case "USBStick":
		mediaKind = rf.USBStickMediaType
	case "DVD":
		mediaKind = rf.DVDMediaType
	default:
		return false, errors.New("invalid media type")
	}

	for _, manager := range managers {
		virtualMedia, err := manager.VirtualMedia()
		if err != nil {
			return false, err
		}
		for _, media := range virtualMedia {
			if media.Inserted {
				err = media.EjectMedia()
				if err != nil {
					return false, err
				}
			}
		}
	}

	if mediaUrl != "" {
		setMedia := false
		for _, manager := range managers {
			virtualMedia, err := manager.VirtualMedia()
			if err != nil {
				return false, err
			}

			for _, media := range virtualMedia {
				for _, t := range media.MediaTypes {
					if t == mediaKind {
						err = media.InsertMedia(mediaUrl, true, true)
						if err != nil {
							return false, err
						}
						setMedia = true
						break
					}
				}
			}
		}
		if !setMedia {
			return false, fmt.Errorf("media kind %s not supported", kind)
		}
	}

	return true, nil
}
