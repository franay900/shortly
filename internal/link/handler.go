package link

import (
	"net/http"
	"strconv"
	"url/short/configs"
	"url/short/pkg/errors"
	"url/short/pkg/event"
	"url/short/pkg/middleware"
	"url/short/pkg/req"
	"url/short/pkg/res"

	"gorm.io/gorm"
)

type LinkHandlerDeps struct {
	LinkRepository *LinkRepository
	Config         *configs.Config
	EventBus       *event.EventBus
}

type LinkHandler struct {
	LinkRepository *LinkRepository
	EventBus       *event.EventBus
}

func NewLinkHandler(router *http.ServeMux, deps LinkHandlerDeps) {

	handler := &LinkHandler{
		LinkRepository: deps.LinkRepository,
		EventBus:       deps.EventBus,
	}
	router.HandleFunc("POST /link", handler.Create())
	router.HandleFunc("GET /link", handler.GetAll())
	router.Handle("PATCH /link/{id}", middleware.IsAuthed(handler.Update(), deps.Config))
	router.HandleFunc("DELETE /link/{id}", handler.Delete())
	router.HandleFunc("GET /{alias}", handler.GoTo())

}

func (handler *LinkHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := req.HandleBody[LinkCreateRequest](w, r)
		if err != nil {
			return
		}

		// Генерируем уникальный хеш с проверкой коллизий
		link, err := handler.generateUniqueLink(body.Url)
		if err != nil {
			errors.WriteError(w, err, errors.GetStatusCode(err))
			return
		}

		createdLink, err := handler.LinkRepository.Create(link)
		if err != nil {
			errors.WriteError(w, errors.WrapError(err, errors.ErrCodeDatabaseError, "Failed to create link"), http.StatusInternalServerError)
			return
		}

		res.Json(w, createdLink, http.StatusCreated)
	}
}

// generateUniqueLink генерирует ссылку с уникальным хешем
func (handler *LinkHandler) generateUniqueLink(url string) (*Link, error) {
	const maxAttempts = 10
	
	for i := 0; i < maxAttempts; i++ {
		link := NewLink(url)
		
		// Проверяем, существует ли уже такой хеш
		existedLink, err := handler.LinkRepository.GetByHash(link.Hash)
		if err != nil {
			// Если это ошибка "record not found", то хеш уникален
			if err.Error() == "record not found" {
				return link, nil
			}
			return nil, errors.WrapError(err, errors.ErrCodeDatabaseError, "Failed to check hash uniqueness")
		}
		
		if existedLink == nil {
			return link, nil
		}
		
		// Если хеш уже существует, генерируем новый
		link.generateHash()
	}
	
	return nil, errors.NewAppError(errors.ErrCodeInternalError, "Failed to generate unique hash after multiple attempts")
}

func (handler *LinkHandler) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := r.Context().Value(middleware.ContextEmailKey).(string)
		if !ok {
			errors.WriteError(w, errors.ErrUnauthorized, http.StatusUnauthorized)
			return
		}
		
		body, err := req.HandleBody[LinkUpdateRequest](w, r)
		if err != nil {
			return
		}

		idString := r.PathValue("id")
		id, err := strconv.ParseInt(idString, 10, 32)
		if err != nil {
			errors.WriteError(w, errors.NewAppError(errors.ErrCodeInvalidParameter, "Invalid link ID"), http.StatusBadRequest)
			return
		}

		link, err := handler.LinkRepository.Update(&Link{
			Model: gorm.Model{ID: uint(id)},
			Url:   body.Url,
			Hash:  body.Hash,
		})
		if err != nil {
			errors.WriteError(w, errors.WrapError(err, errors.ErrCodeDatabaseError, "Failed to update link"), http.StatusInternalServerError)
			return
		}

		res.Json(w, link, http.StatusOK)
	}
}

func (handler *LinkHandler) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idString := r.PathValue("id")
		id, err := strconv.ParseInt(idString, 10, 32)
		if err != nil {
			errors.WriteError(w, errors.NewAppError(errors.ErrCodeInvalidParameter, "Invalid link ID"), http.StatusBadRequest)
			return
		}

		_, err = handler.LinkRepository.GetById(uint(id))
		if err != nil {
			errors.WriteError(w, errors.ErrNotFound, http.StatusNotFound)
			return
		}

		err = handler.LinkRepository.Delete(uint(id))
		if err != nil {
			errors.WriteError(w, errors.WrapError(err, errors.ErrCodeDatabaseError, "Failed to delete link"), http.StatusInternalServerError)
			return
		}
		
		res.Json(w, nil, http.StatusOK)
	}
}

func (handler *LinkHandler) GoTo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hash := r.PathValue("alias")
		link, err := handler.LinkRepository.GetByHash(hash)
		if err != nil {
			errors.WriteError(w, errors.ErrNotFound, http.StatusNotFound)
			return
		}

		go handler.EventBus.Publish(event.Event{
			Type: event.EventLinkVisited,
			Data: link.ID,
		})
		http.Redirect(w, r, link.Url, http.StatusTemporaryRedirect)
	}
}

func (handler *LinkHandler) GetAll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Валидация и парсинг параметров с дефолтными значениями
		limitStr := r.URL.Query().Get("limit")
		if limitStr == "" {
			limitStr = "10" // Дефолтное значение
		}
		
		offsetStr := r.URL.Query().Get("offset")
		if offsetStr == "" {
			offsetStr = "0" // Дефолтное значение
		}
		
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 1000 {
			errors.WriteError(w, errors.NewAppError(errors.ErrCodeInvalidParameter, "Limit must be a positive integer between 1 and 1000"), http.StatusBadRequest)
			return
		}

		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			errors.WriteError(w, errors.NewAppError(errors.ErrCodeInvalidParameter, "Offset must be a non-negative integer"), http.StatusBadRequest)
			return
		}

		links := handler.LinkRepository.Get(limit, offset)
		count := handler.LinkRepository.Count()

		res.Json(w, &GetAllLinksResponse{
			Links: links,
			Count: count,
		}, http.StatusOK)
	}
}
