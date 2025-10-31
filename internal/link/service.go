package link

import (
    "errors"
    "url/short/pkg/event"

    "gorm.io/gorm"
)

type LinkService struct {
    repo     *LinkRepository
    eventBus *event.EventBus
}

func NewLinkService(repo *LinkRepository, eventBus *event.EventBus) *LinkService {
    return &LinkService{repo: repo, eventBus: eventBus}
}

// Create generates a unique hash and persists the link.
func (s *LinkService) Create(url string) (*Link, error) {
    link := NewLink(url)

    // ensure uniqueness of hash
    for {
        existed, _ := s.repo.GetByHash(link.Hash)
        if existed == nil {
            break
        }
        link.generateHash()
    }

    created, err := s.repo.Create(link)
    if err != nil {
        return nil, err
    }
    return created, nil
}

// Update updates link fields by id.
func (s *LinkService) Update(id uint, url, hash string) (*Link, error) {
    // Optional: ensure hash uniqueness if provided
    if hash != "" {
        existed, _ := s.repo.GetByHash(hash)
        if existed != nil && existed.ID != id {
            return nil, errors.New("hash already in use")
        }
    }

    link, err := s.repo.Update(&Link{Model: gorm.Model{ID: id}, Url: url, Hash: hash})
    if err != nil {
        return nil, err
    }
    return link, nil
}

// Delete removes link by id.
func (s *LinkService) Delete(id uint) error {
    return s.repo.Delete(id)
}

func (s *LinkService) GetByID(id uint) (*Link, error) {
    return s.repo.GetById(id)
}

// GetAll returns paginated list and total count.
func (s *LinkService) GetAll(limit, offset int) ([]Link, int64) {
    links := s.repo.Get(limit, offset)
    count := s.repo.Count()
    return links, count
}

// Visit finds link by alias and publishes event.
func (s *LinkService) Visit(alias string) (*Link, error) {
    link, err := s.repo.GetByHash(alias)
    if err != nil {
        return nil, err
    }
    go s.eventBus.Publish(event.Event{Type: event.EventLinkVisited, Data: link.ID})
    return link, nil
}

// no extra helpers