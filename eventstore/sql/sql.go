package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/eventstore"
)

// SQL for store events
type SQL struct {
	db         *sql.DB
	serializer eventsourcing.Serializer
}

// Open connection to database
func Open(db *sql.DB, serializer eventsourcing.Serializer) *SQL {
	return &SQL{
		db:         db,
		serializer: serializer,
	}
}

// Close the connection
func (s *SQL) Close() {
	s.db.Close()
}

// Save persists events to the database
func (s *SQL) Save(events []eventsourcing.Event) error {
	// If no event return no error
	if len(events) == 0 {
		return nil
	}
	aggregateID := events[0].AggregateID
	aggregateType := events[0].AggregateType

	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return errors.New(fmt.Sprintf("could not start a write transaction, %v", err))
	}
	defer tx.Rollback()

	var currentVersion eventsourcing.Version
	var version int
	selectStm := `Select version from events where id=? and type=? order by version desc limit 1`
	err = tx.QueryRow(selectStm, aggregateID, aggregateType).Scan(&version)
	if err != nil && err != sql.ErrNoRows {
		return err
	} else if err == sql.ErrNoRows {
		// if no events are saved before set the current version to zero
		currentVersion = eventsourcing.Version(0)
	} else {
		// set the current version to the last event stored
		currentVersion = eventsourcing.Version(version)
	}

	//Validate events
	err = eventstore.ValidateEvents(aggregateID, currentVersion, events)
	if err != nil {
		return err
	}

	insert := `Insert into events (id, version, reason, type, timestamp, data, metadata) values ($1, $2, $3, $4, $5, $6, $7)`
	for _, event := range events {
		var e, m []byte

		e, err := s.serializer.Marshal(event.Data)
		if err != nil {
			return err
		}
		if event.MetaData != nil {
			m, err = s.serializer.Marshal(event.MetaData)
			if err != nil {
				return err
			}
		}
		_, err = tx.Exec(insert, event.AggregateID, event.Version, event.Reason, event.AggregateType, event.Timestamp.Format(time.RFC3339), string(e), string(m))
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Get the events from database
func (s *SQL) Get(id string, aggregateType string, afterVersion eventsourcing.Version) (events []eventsourcing.Event, err error) {
	selectStm := `Select id, version, reason, type, timestamp, data, metadata from events where id=? and type=? and version>? order by version asc`
	rows, err := s.db.Query(selectStm, id, aggregateType, afterVersion)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var eventMetaData map[string]interface{}
		var version eventsourcing.Version
		var id, reason, typ, timestamp string
		var data, metadata string
		if err := rows.Scan(&id, &version, &reason, &typ, &timestamp, &data, &metadata); err != nil {
			return nil, err
		}

		t, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			return nil, err
		}

		f, ok := s.serializer.Type(typ, reason)
		if !ok {
			// if the typ/reason is not register jump over the event
			continue
		}

		eventData := f()
		err = s.serializer.Unmarshal([]byte(data), &eventData)
		if err != nil {
			return nil, err
		}
		if metadata != "" {
			err = s.serializer.Unmarshal([]byte(metadata), &eventMetaData)
			if err != nil {
				return nil, err
			}
		}

		events = append(events, eventsourcing.Event{
			AggregateID:   id,
			Version:       version,
			AggregateType: typ,
			Reason:        reason,
			Timestamp:     t,
			Data:          eventData,
			MetaData:      eventMetaData,
		})
	}
	return
}
