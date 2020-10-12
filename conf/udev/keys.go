package udev

const (
	// MatchKernel matches on the kernel device name.
	MatchKernel = "KERNEL"
	// MatchBus matches the bus type of the device.
	MatchBus = "BUS"
	// MatchID matches the device number on the bus, ie: PCI bus ID.
	MatchID = "ID"
	// MatchSubsystem matches the subsystem of the device.
	MatchSubsystem = "SUBSYSTEM"
	// MatchPlace matches the position on the bus.
	MatchPlace = "PLACE"
	// MatchSysfsLabel matches the label attribute in sysfs.
	MatchSysfsLabel = "SYSFS{label}"
	// MatchSysfsSerial matches the serial attribute in sysfs.
	MatchSysfsSerial = "SYSFS{serial}"
	// MatchSysfsVendor matches the vendor attribute in sysfs.
	MatchSysfsVendor = "SYSFS{vendor}"
	// MatchAnyProductID matches the productId of a device or any of its parents.
	MatchAnyProductID = "ATTRS{idProduct}"
	// MatchAnyVendorID matches the vendorId of a device or any of its parents.
	MatchAnyVendorID = "ATTRS{idVendor}"
)

const (
	// ActionSymlink creates a symlink for matched devices.
	ActionSymlink = "SYMLINK"
	// ActionName sets the name for matched device nodes (to be created).
	ActionName = "NAME"
	// ActionOwner sets the node owner.
	ActionOwner = "OWNER"
	// ActionGroup sets the node group.
	ActionGroup = "GROUP"
	// ActionMode sets the node mode.
	ActionMode = "MODE"
	// ActionEnv sets a device property value.
	ActionEnv = "{ENV}"
)
