package dg

import (
    "fmt"
    "encoding/xml"

    "github.com/PaloAltoNetworks/pango/util"
)

// Dg is the client.Panorama.DeviceGroup namespace.
type Dg struct {
    con util.XapiClient
}

// Initialize is invoked by client.Initialize().
func (c *Dg) Initialize(con util.XapiClient) {
    c.con = con
}

/*
AddDevices performs a SET to add devices to device group g.

The device group can be either a string or an Entry object.
*/
func (c *Dg) AddDevices(g interface{}, e ...string) error {
    var name string

    switch v := g.(type) {
    case string:
        name = v
    case Entry:
        name = v.Name
    default:
        return fmt.Errorf("Unknown type sent to add devices: %s", v)
    }

    c.con.LogAction("(set) devices in device group: %s", name)

    ent := util.StrToEnt(e)
    if ent == nil {
        return nil
    }

    // Set xpath.
    path := c.xpath([]string{name})
    if len(ent.Entries) == 1 {
        path = append(path, "devices")
    }

    dv := make([]interface{}, len(ent.Entries))
    for i := range ent.Entries {
        dv[i] = ent.Entries[i]
    }

    d := util.BulkElement{XMLName: xml.Name{Local: "devices"}, Data: dv}

    _, err := c.con.Set(path, d.Config(), nil, nil)
    return err
}

/*
DeleteDevices performs a DELETE to remove devices d from device group g.

The device group can be either a string or an Entry object.
*/
func (c *Dg) DeleteDevices(g interface{}, d ...string) error {
    var name string

    switch v := g.(type) {
    case string:
        name = v
    case Entry:
        name = v.Name
    default:
        return fmt.Errorf("Unknown type sent to remove devices: %s", v)
    }

    c.con.LogAction("(delete) devices in device group: %s", name)

    path := c.xpath([]string{name})
    path = append(path, "devices", util.AsEntryXpath(d))

    _, err := c.con.Delete(path, nil, nil)
    return err
}

// ShowList performs SHOW to retrieve a list of device groups.
func (c *Dg) ShowList() ([]string, error) {
    c.con.LogQuery("(show) list of device groups")
    path := c.xpath(nil)
    return c.con.EntryListUsing(c.con.Show, path[:len(path) - 1])
}

// GetList performs GET to retrieve a list of device groups.
func (c *Dg) GetList() ([]string, error) {
    c.con.LogQuery("(get) list of device groups")
    path := c.xpath(nil)
    return c.con.EntryListUsing(c.con.Get, path[:len(path) - 1])
}

// Get performs GET to retrieve information for the given device group.
func (c *Dg) Get(name string) (Entry, error) {
    c.con.LogQuery("(get) device group %q", name)
    return c.details(c.con.Get, name)
}

// Show performs SHOW to retrieve information for the given device group.
func (c *Dg) Show(name string) (Entry, error) {
    c.con.LogQuery("(show) device group %q", name)
    return c.details(c.con.Show, name)
}

// Set performs SET to create / update one or more device groups.
func (c *Dg) Set(e ...Entry) error {
    var err error

    if len(e) == 0 {
        return nil
    }

    _, fn := c.versioning()
    names := make([]string, len(e))

    // Build up the struct with the given configs.
    d := util.BulkElement{XMLName: xml.Name{Local: "device-group"}}
    for i := range e {
        d.Data = append(d.Data, fn(e[i]))
        names[i] = e[i].Name
    }
    c.con.LogAction("(set) device groups: %v", names)

    // Set xpath.
    path := c.xpath(names)
    if len(e) == 1 {
        path = path[:len(path) - 1]
    } else {
        path = path[:len(path) - 2]
    }

    // Create the device groups.
    _, err = c.con.Set(path, d.Config(), nil, nil)
    return err
}

// Edit performs EDIT to create / update a device group.
func (c *Dg) Edit(e Entry) error {
    var err error

    _, fn := c.versioning()

    c.con.LogAction("(edit) device group %q", e.Name)

    // Set xpath.
    path := c.xpath([]string{e.Name})

    // Edit the device group.
    _, err = c.con.Edit(path, fn(e), nil, nil)
    return err
}

// Delete removes the given device groups from the firewall.
//
// Device groups can be a string or an Entry object.
func (c *Dg) Delete(e ...interface{}) error {
    var err error

    if len(e) == 0 {
        return nil
    }

    names := make([]string, len(e))
    for i := range e {
        switch v := e[i].(type) {
        case string:
            names[i] = v
        case Entry:
            names[i] = v.Name
        default:
            return fmt.Errorf("Unknown type sent to delete: %s", v)
        }
    }
    c.con.LogAction("(delete) device groups: %v", names)

    // Remove the device groups.
    path := c.xpath(names)
    _, err = c.con.Delete(path, nil, nil)
    return err
}

/** Internal functions for the Dg struct **/

func (c *Dg) versioning() (normalizer, func(Entry) (interface{})) {
    return &container_v1{}, specify_v1
}

func (c *Dg) details(fn util.Retriever, name string) (Entry, error) {
    path := c.xpath([]string{name})
    obj, _ := c.versioning()
    if _, err := fn(path, nil, obj); err != nil {
        return Entry{}, err
    }
    ans := obj.Normalize()

    return ans, nil
}

func (c *Dg) xpath(vals []string) []string {
    return []string{
        "config",
        "devices",
        util.AsEntryXpath([]string{"localhost.localdomain"}),
        "device-group",
        util.AsEntryXpath(vals),
    }
}
