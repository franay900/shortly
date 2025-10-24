package stat

import (
	"fmt"
	"url/short/pkg/event"
)

type StatServiceDeps struct {
	EventBus       *event.EventBus
	StatRepository *StatRepository
}

type StatService struct {
	EventBus       *event.EventBus
	StatRepository *StatRepository
}

func NewStatService(deps *StatServiceDeps) *StatService {
	return &StatService{
		EventBus:       deps.EventBus,
		StatRepository: deps.StatRepository,
	}
}

func (s *StatService) AddClick() {
	for {
		select {
		case msg := <-s.EventBus.Subscribe():
			if msg.Type == event.EventLinkVisited {
				s.StatRepository.AddClick(msg.Data.(uint))
				fmt.Println("Event link visited")
			}

		}
	}
}
