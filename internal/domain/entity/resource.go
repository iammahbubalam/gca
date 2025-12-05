package entity

// Resource represents system resource information
type Resource struct {
	TotalCPU      int // Total CPU cores
	AvailableCPU  int // Available CPU cores
	ReservedCPU   int // Reserved CPU cores for owner
	
	TotalRAMGB    int // Total RAM in GB
	AvailableRAMGB int // Available RAM in GB
	ReservedRAMGB int // Reserved RAM for owner
	
	TotalDiskGB    int // Total disk space in GB
	AvailableDiskGB int // Available disk space in GB
	ReservedDiskGB int // Reserved disk space for owner
}

// CanAllocate checks if resources can be allocated for a VM
func (r *Resource) CanAllocate(vcpu, ramGB, diskGB int) bool {
	return r.AvailableCPU >= vcpu &&
		r.AvailableRAMGB >= ramGB &&
		r.AvailableDiskGB >= diskGB
}

// Allocate reduces available resources
func (r *Resource) Allocate(vcpu, ramGB, diskGB int) {
	r.AvailableCPU -= vcpu
	r.AvailableRAMGB -= ramGB
	r.AvailableDiskGB -= diskGB
}

// Release increases available resources
func (r *Resource) Release(vcpu, ramGB, diskGB int) {
	r.AvailableCPU += vcpu
	r.AvailableRAMGB += ramGB
	r.AvailableDiskGB += diskGB
}
