package eventsourcing_test

import (
	"github.com/hallgren/eventsourcing"
	"testing"
)

type AnEvent struct {
	Name string
}

type AnotherEvent struct{}

var event = eventsourcing.Event{Version: 123, Data: &AnEvent{Name: "123"}}
var otherEvent = eventsourcing.Event{Version: 123, Data: &AnotherEvent{}}

func TestGlobal(t *testing.T) {
	var streamEvent *eventsourcing.Event
	e := eventsourcing.NewEventStream()
	f := func(e eventsourcing.Event) {
		streamEvent = &e
	}
	e.Subscribe(f)
	e.Update(event)

	//time.Sleep(1 * time.Second)
	if streamEvent == nil {
		t.Fatalf("should have received event")
	}
	if streamEvent.Version != event.Version {
		t.Fatalf("wrong info in event got %q expected %q", streamEvent.Version, event.Version)
	}
}


func TestSpecific(t *testing.T) {
	var streamEvent *eventsourcing.Event
	e := eventsourcing.NewEventStream()
	f := func(e eventsourcing.Event) {
		streamEvent = &e
	}
	e.Subscribe(f,&AnEvent{})
	e.Update(event)

	if streamEvent == nil {
		t.Fatalf("should have received event")
	}

	if streamEvent.Version != event.Version {
		t.Fatalf("wrong info in event got %q expected %q", streamEvent.Version, event.Version)
	}
}

func TestManySpecific(t *testing.T) {
	var streamEvents []*eventsourcing.Event
	e := eventsourcing.NewEventStream()
	f := func(e eventsourcing.Event) {
		streamEvents = append(streamEvents, &e)
	}
	e.Subscribe(f, &AnEvent{}, &AnotherEvent{})
	e.Update(event)
	e.Update(otherEvent)

	if streamEvents == nil {
		t.Fatalf("should have received event")
	}

	if len(streamEvents) != 2 {
		t.Fatalf("should have received 2 events")
	}

	switch ev := streamEvents[0].Data.(type) {
	case *AnotherEvent:
		t.Fatalf("expecting AnEvent got %q", ev)
	}

	switch ev := streamEvents[1].Data.(type) {
	case *AnEvent:
		t.Fatalf("expecting OtherEvent got %q", ev)
	}

}

func TestUpdateNoneSubscribedEvent(t *testing.T) {
	var streamEvent *eventsourcing.Event
	e := eventsourcing.NewEventStream()
	f := func(e eventsourcing.Event) {
		streamEvent = &e
	}
	e.Subscribe(f, &AnotherEvent{})
	e.Update(event)

	if streamEvent != nil {
		t.Fatalf("should not have received event %q", streamEvent)
	}
}

func TestManySubscribers(t *testing.T) {
	streamEvent1 := make([]eventsourcing.Event,0)
	streamEvent2 := make([]eventsourcing.Event,0)
	streamEvent3 := make([]eventsourcing.Event,0)
	streamEvent4 := make([]eventsourcing.Event,0)

	e := eventsourcing.NewEventStream()
	f1 := func(e eventsourcing.Event) {
		streamEvent1 = append(streamEvent1,e)
	}
	f2 := func(e eventsourcing.Event) {
		streamEvent2 = append(streamEvent2, e)
	}
	f3 := func(e eventsourcing.Event) {
		streamEvent3 = append(streamEvent3, e)
	}
	f4 := func(e eventsourcing.Event) {
		streamEvent4 = append(streamEvent4,e)
	}
	e.Subscribe(f1,&AnotherEvent{})
	e.Subscribe(f2,&AnotherEvent{}, &AnEvent{})
	e.Subscribe(f3,&AnEvent{})
	e.Subscribe(f4)

	e.Update(event)

	if len(streamEvent1) != 0 {
		t.Fatalf("stream1 should not have any events")
	}

	if len(streamEvent2) != 1 {
		t.Fatalf("stream2 should have one event")
	}

	if len(streamEvent3) != 1 {
		t.Fatalf("stream3 should have one event")
	}

	if len(streamEvent4) != 1 {
		t.Fatalf("stream4 should have one event")
	}
}
