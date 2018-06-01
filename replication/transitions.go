package replication

type transitionToMaster struct{}

type transitionToSlave struct {
	MasterAddr string
}
