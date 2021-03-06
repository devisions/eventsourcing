package memory

import (
	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/snapshotstore"
)

// Handler of snapshot store
type Handler struct {
	store      map[string][]byte
	serializer eventsourcing.Serializer
}

// New handler for the snapshot service
func New(serializer eventsourcing.Serializer) *Handler {
	return &Handler{
		store:      make(map[string][]byte),
		serializer: serializer,
	}
}

// Get returns the deserialize snapshot
func (h *Handler) Get(id string, s eventsourcing.Aggregate) error {
	v, ok := h.store[id]
	if !ok {
		return eventsourcing.ErrSnapshotNotFound
	}
	return h.serializer.Unmarshal(v, s)
}

// Save persists the snapshot
func (h *Handler) Save(a eventsourcing.Aggregate) error {
	root := a.Root()
	err := snapshotstore.Validate(*root)
	if err != nil {
		return err
	}

	data, err := h.serializer.Marshal(a)
	if err != nil {
		return err
	}

	h.store[root.ID()] = data
	return nil
}
