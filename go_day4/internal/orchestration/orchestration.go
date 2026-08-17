package orchestration

import (
	"context"
	"fmt"
)

type Service interface {
	Name() string
	Start(ctx context.Context) error
}

func StartServices(ctx context.Context, services []Service) error{
	for _, service := range services{
		if err := service.Start(ctx); err != nil{
			return fmt.Errorf("start %s: %w", service.Name(), err)
		}
	}
	return nil
}