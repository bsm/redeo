package replication

import (
	"github.com/bsm/redeo/internal/util"
)

type runtimeState struct {
	ID      string `json:"id"`
	Offset  int64  `json:"offset"`
	SlaveOf string `json:"slave_of,omitempty"`
}

func (s *runtimeState) Load(store StableStore) error {
	err := store.Decode("replication", s)
	if err == ErrNotFound {
		s.ID = util.GenerateID(20)
		return nil
	}
	return err
}

func (s *runtimeState) Save(store StableStore) error {
	return store.Encode("replication", s)
}
