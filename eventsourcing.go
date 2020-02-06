package eventsourcing

import (
	"errors"
	"reflect"
	"time"

	uuid "github.com/gofrs/uuid"
)

// Version is the event version used in event and aggregateRoot
type Version int

// AggregateRoot to be included into aggregates
type AggregateRoot struct {
	AggregateID      string
	AggregateVersion Version
	AggregateEvents  []Event
}

// Event holding meta data and the application specific event in the Data property
type Event struct {
	AggregateRootID string
	Version         Version
	Reason          string
	AggregateType   string
	Timestamp       time.Time
	Data            interface{}
	MetaData        map[string]interface{}
}

var (
	// ErrAggregateAlreadyExists returned if the AggregateID is set more than one time
	ErrAggregateAlreadyExists = errors.New("its not possible to set id on already existing aggregate")

	emptyAggregateID = ""
)

func (e *Event) aggregateID() string {
	return e.AggregateRootID
}

type event interface {
	aggregateID() string
}

// TrackChange is used internally by behaviour methods to apply a state change to
// the current instance and also track it in order that it can be persisted later.
func (state *AggregateRoot) TrackChange(a aggregate, e event) error {
	return state.TrackChangeWithMetaData(a, e, nil)
}

// TrackChangeWithMetaData is used internally by behaviour methods to apply a state change to
// the current instance and also track it in order that it can be persisted later.
// meta data is handled by this func to store none related application state
func (state *AggregateRoot) TrackChangeWithMetaData(a aggregate, e event, metaData map[string]interface{}) error {
	// This can be overwritten in the constructor of the aggregate
	if state.AggregateID == emptyAggregateID {
		state.setID(uuid.Must(uuid.NewV4()).String())
	}

	reason := reflect.TypeOf(e).Elem().Name()
	aggregateType := reflect.TypeOf(a).Elem().Name()
	event := Event{
		AggregateRootID: state.AggregateID,
		Version:         state.nextVersion(),
		Reason:          reason,
		AggregateType:   aggregateType,
		Timestamp:       time.Now().UTC(),
		Data:            e,
		MetaData:        metaData,
	}
	state.AggregateEvents = append(state.AggregateEvents, event)
	a.Transition(event)
	return nil
}

// BuildFromHistory builds the aggregate state from events
func (state *AggregateRoot) BuildFromHistory(a aggregate, events []Event) {
	for _, event := range events {
		a.Transition(event)
		//Set the aggregate id
		state.AggregateID = event.AggregateRootID
		// Make sure the aggregate is in the correct version (the last event)
		state.AggregateVersion = event.Version
	}
}

func (state *AggregateRoot) nextVersion() Version {
	return state.CurrentVersion() + 1
}

// updateVersion sets the AggregateVersion to the AggregateVersion in the last event if reset the events
// called by the Save func in the repository after the events are stored
func (state *AggregateRoot) updateVersion() {
	if len(state.AggregateEvents) > 0 {
		state.AggregateVersion = state.AggregateEvents[len(state.AggregateEvents)-1].Version
		state.AggregateEvents = []Event{}
	}
}

func (state *AggregateRoot) changes() []Event {
	return state.AggregateEvents
}

// setID is the internal method to set the aggregate id
func (state *AggregateRoot) setID(id string) {
	state.AggregateID = id
}

func (state *AggregateRoot) version() Version {
	return state.AggregateVersion
}

//Public accessors for aggregate root properties

// SetID opens up the possibility to set manual aggregate id from the outside
func (state *AggregateRoot) SetID(id string) error {
	if state.AggregateID != emptyAggregateID {
		return ErrAggregateAlreadyExists
	}
	state.setID(id)
	return nil
}

// ID returns the aggregate id as a string
func (state *AggregateRoot) id() string {
	return state.AggregateID
}

// CurrentVersion return the version based on events that are not stored
func (state *AggregateRoot) CurrentVersion() Version {
	if len(state.AggregateEvents) > 0 {
		return state.AggregateEvents[len(state.AggregateEvents)-1].Version
	}
	return state.AggregateVersion
}
