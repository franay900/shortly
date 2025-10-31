package link

import (
    "net/http"
    "strconv"
    "url/short/configs"
    "url/short/pkg/middleware"
    "url/short/pkg/req"
    "url/short/pkg/res"
)

type LinkHandlerDeps struct {
    LinkService *LinkService
    Config      *configs.Config
}

type LinkHandler struct {
    LinkService *LinkService
}

func NewLinkHandler(router *http.ServeMux, deps LinkHandlerDeps) {

    handler := &LinkHandler{
        LinkService: deps.LinkService,
    }
    router.HandleFunc("POST /link", handler.Create())
    router.HandleFunc("GET /link", handler.GetAll())
    router.Handle("PATCH /link/{id}", middleware.IsAuthed(handler.Update(), deps.Config))
    router.HandleFunc("DELETE /link/{id}", handler.Delete())
    router.HandleFunc("GET /{alias}", handler.GoTo())

}

func (handler *LinkHandler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := req.HandleBody[LinkCreateRequest](&w, r)

		if err != nil {
			return
		}

    createdLink, err := handler.LinkService.Create(body.Url)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    res.Json(w, createdLink, http.StatusCreated)

	}

}

func (handler *LinkHandler) Update() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
    body, err := req.HandleBody[LinkUpdateRequest](&w, r)
    if err != nil {
        return
    }

    idString := r.PathValue("id")
    id, err := strconv.ParseInt(idString, 10, 32)

    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    link, err := handler.LinkService.Update(uint(id), body.Url, body.Hash)

    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
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
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // ensure exists
    _, err = handler.LinkService.GetByID(uint(id))
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    err = handler.LinkService.Delete(uint(id))

    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusNoContent)

	}
}

func (handler *LinkHandler) GoTo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
    hash := r.PathValue("alias")
    link, err := handler.LinkService.Visit(hash)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    http.Redirect(w, r, link.Url, http.StatusTemporaryRedirect)
	}
}

func (handler *LinkHandler) GetAll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
    limitStr := r.URL.Query().Get("limit")
    offsetStr := r.URL.Query().Get("offset")

    limit := 10
    offset := 0

    if limitStr != "" {
        l, err := strconv.Atoi(limitStr)
        if err != nil || l < 0 {
            http.Error(w, "Error with parsing limit", http.StatusBadRequest)
            return
        }
        limit = l
    }

    if offsetStr != "" {
        o, err := strconv.Atoi(offsetStr)
        if err != nil || o < 0 {
            http.Error(w, "Error with parsing offset", http.StatusBadRequest)
            return
        }
        offset = o
    }

    links, count := handler.LinkService.GetAll(limit, offset)

    res.Json(w, &GetAllLinksResponse{
        Links: links,
        Count: count,
    }, http.StatusOK)
	}
}
