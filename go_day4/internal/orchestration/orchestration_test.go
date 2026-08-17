package orchestration_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"honeypot-day4/internal/orchestration"
)

type fakeService struct {
	name string
	started *[]string
	err error
}

func (s fakeService) Name() string {
	return s.name
}

func (s fakeService) Start(context.Context) error{
	*s.started = append(*s.started, s.name)
	return s.err
}

func TestStartServicesUsesOnlyServiceContract(t *testing.T){
	var started []string
	services := []orchestration.Service{
		fakeService{name: "telnet", started: &started},
		fakeService{name: "http", started: &started},
	}
	if err := orchestration.StartServices(context.Background(), services); err != nil {
		t.Fatalf("StartServices() error = %v", err)
	}	
	want := []string{"telnet", "http"}
	if !reflect.DeepEqual(started, want){
		t.Fatalf("started = %v, want %v", started, want)
	}
}

func TestStartServicesStopsAndWrapsError(t *testing.T){
	var started []string
	startErr := errors.New("port unavailable")
	services := []orchestration.Service{
		fakeService{name: "telnet", started: &started, err: startErr},
		fakeService{name: "http", started: &started},
	}

	err := orchestration.StartServices(context.Background(), services)
	if !errors.Is(err, startErr){
		t.Fatalf("StartServices() error = %v, want wrapped %v", err, startErr)
	}
	if got, want := err.Error(), "start telnet: port unavailable"; got != want{
		t.Fatalf("StartServices() error = %q, want %q", got, want)
	}

	want := []string{"telnet"}
	if !reflect.DeepEqual(started, want){
		t.Fatalf("started = %v, want %v", started, want)
	}
}