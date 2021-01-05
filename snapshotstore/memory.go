package snapshotstore

import (
	"errors"
	"github.com/hallgren/eventsourcing"
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
func (h *Handler) Get(id string, s eventsourcing.Snapshot) error {
	v, ok := h.store[id]
	if !ok {
		return eventsourcing.ErrSnapshotNotFound
	}
	err := h.serializer.Unmarshal(v, s)
	if err != nil {
		return err
	}
	return nil
}

// Save persists the snapshot
func (h *Handler) Save(s eventsourcing.Snapshot) error {
	if s.ID() == "" {
		return errors.New("aggregate id is empty")
	}
	if s.UnsavedEvents() {
		return errors.New("aggregate holds unsaved events")
	}
	data, err := h.serializer.Marshal(s)
	if err != nil {
		return err
	}
	h.store[s.ID()] = data
	return nil
}
