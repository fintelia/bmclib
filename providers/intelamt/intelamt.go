package intelamt

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/bmc-toolbox/bmclib/v2/providers"
	"github.com/go-logr/logr"
	"github.com/jacobweinstock/go-amt"
	"github.com/jacobweinstock/registrar"
)

const (
	// ProviderName for the provider AMT implementation
	ProviderName = "IntelAMT"
	// ProviderProtocol for the provider AMT implementation
	ProviderProtocol = "AMT"
)

var (
	// Features implemented by the AMT provider
	Features = registrar.Features{
		providers.FeaturePowerSet,
		providers.FeaturePowerState,
		providers.FeatureBootDeviceSet,
	}
)

type amtProvider interface {
	IsPoweredOn(context.Context) (bool, error)
	PowerOn(context.Context) error
	PowerOff(context.Context) error
	PowerCycle(context.Context) error
	SetPXE(context.Context) error
	Close() error
}

// Conn is a connection to a BMC via AMT
type Conn struct {
	Host   string
	Port   uint32
	User   string
	Pass   string
	Log    logr.Logger
	client amtProvider
}

// New creates a new AMT connection
func New(log logr.Logger, host string, port string, user string, pass string) *Conn {
	p, err := strconv.Atoi(port)
	if err != nil {
		p = 16992
	}
	conn := amt.Connection{
		Host:   host,
		Port:   uint32(p),
		User:   user,
		Pass:   pass,
		Logger: log,
	}
	client, _ := amt.NewClient(conn)
	c := &Conn{
		Host:   host,
		Port:   uint32(p),
		User:   user,
		Pass:   pass,
		Log:    log,
		client: client,
	}

	return c
}

// Name of the provider
func (c *Conn) Name() string {
	return ProviderName
}

// Open a connection to the BMC via AMT.
// The AMT library does not do/use sessions so opening just instantiates the Conn.client.
// It will communicate with the BMC.
func (c *Conn) Open(ctx context.Context) (err error) {
	if c.client == nil {
		conn := amt.Connection{
			Host:   c.Host,
			Port:   c.Port,
			User:   c.User,
			Pass:   c.Pass,
			Logger: c.Log,
		}
		client, err := amt.NewClient(conn)
		if err != nil {
			return err
		}
		c.client = client
	}

	return nil
}

// Close a connection to a BMC
func (c *Conn) Close() (err error) {
	return c.client.Close()
}

// Compatible tests whether a BMC is compatible with the ipmitool provider
func (c *Conn) Compatible(ctx context.Context) bool {
	amtclient, ok := c.client.(*amt.Client)
	if !ok || amtclient == nil {
		return false
	}

	if _, err := amtclient.IsPoweredOn(ctx); err != nil {
		return false
	}

	return true
}

// BootDeviceSet sets the next boot device with options
func (c *Conn) BootDeviceSet(ctx context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
	if strings.ToLower(bootDevice) != "pxe" {
		return false, errors.New("only pxe boot device is supported for AMT provider")
	}
	if err := c.client.SetPXE(ctx); err != nil {
		return false, err
	}

	return true, nil
}

// PowerStateGet gets the power state of a BMC machine
func (c *Conn) PowerStateGet(ctx context.Context) (state string, err error) {
	on, err := c.client.IsPoweredOn(ctx)
	if err != nil {
		return "", err
	}
	if !on {
		return "off", nil
	}

	return "on", nil
}

// PowerSet sets the power state of a BMC machine
func (c *Conn) PowerSet(ctx context.Context, state string) (ok bool, err error) {
	on, _ := c.client.IsPoweredOn(ctx)

	switch strings.ToLower(state) {
	case "on":
		if on {
			return true, nil
		}
		if err := c.client.PowerOn(ctx); err != nil {
			return false, err
		}
		ok = true
	case "off":
		if !on {
			return true, nil
		}
		if err := c.client.PowerOff(ctx); err != nil {
			return false, err
		}
		ok = true
	case "cycle":
		if err := c.client.PowerCycle(ctx); err != nil {
			return false, err
		}
		ok = true
	default:
		err = errors.New("requested state type unknown")
	}

	return ok, err
}
