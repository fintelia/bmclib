package bmclib

import (
	"context"
	"testing"
	"time"

	"github.com/bmc-toolbox/bmclib/v2/logging"
)

func TestBMC(t *testing.T) {
	t.Skip("needs ipmitool and real ipmi server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	host := "127.0.0.1"
	port := "623"
	user := "ADMIN"
	pass := "ADMIN"

	log := logging.DefaultLogger()
	cl := NewClient(host, port, user, pass, WithLogger(log))
	cl.Registry.Drivers = cl.Registry.FilterForCompatible(ctx)
	var err error
	err = cl.Open(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer cl.Close(ctx)
	t.Logf("metadata: %+v", cl.GetMetadata())

	cl.Registry.Drivers = cl.Registry.PreferDriver("dummy")
	state, err := cl.GetPowerState(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(state)
	t.Logf("metadata %+v", cl.GetMetadata())

	cl.Registry.Drivers = cl.Registry.PreferDriver("ipmitool")
	state, err = cl.GetPowerState(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(state)
	t.Logf("metadata: %+v", cl.GetMetadata())

	users, err := cl.ReadUsers(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(users)
	t.Logf("metadata: %+v", cl.GetMetadata())

	t.Fatal()
}
