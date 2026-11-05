package devhandler

import (
	"net/http"

	"github.com/Jarimnark/sugar-cane-service/internal/domain"
	"github.com/Jarimnark/sugar-cane-service/internal/httputil"
	appmw "github.com/Jarimnark/sugar-cane-service/internal/middleware"
)

const (
	maxMediaSize     = 2 << 20
	screenshotPrefix = "screenshots"
	bannerPrefix     = "banners"
)

var allowedMediaTypes = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/webp": {},
}

// UploadScreenshot handles POST /me/game/screenshot?kind=banner|screenshot.
// Both media kinds share the 2 MB / jpeg|png|webp constraints. Banner kind
// writes to the banners/ S3 prefix; screenshot kind to screenshots/.
func (h *Handler) UploadScreenshot(w http.ResponseWriter, r *http.Request) {
	if _, ok := appmw.UserIDFromCtx(r.Context()); !ok {
		httputil.RespondError(w, domain.Unauthorized("not authenticated"))
		return
	}
	if h.storage == nil {
		httputil.RespondError(w, domain.Internal(nil))
		return
	}

	kind := r.URL.Query().Get("kind")
	prefix := screenshotPrefix
	switch kind {
	case "", "screenshot":
		prefix = screenshotPrefix
	case "banner":
		prefix = bannerPrefix
	default:
		httputil.RespondError(w, domain.BadRequest("invalid kind; use banner or screenshot", nil))
		return
	}

	if err := r.ParseMultipartForm(maxMediaSize + 1024); err != nil {
		httputil.RespondError(w, domain.BadRequest("invalid multipart form", err))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		httputil.RespondError(w, domain.BadRequest("missing file field", err))
		return
	}
	defer file.Close()

	if header.Size > maxMediaSize {
		httputil.RespondError(w, domain.BadRequest("file exceeds 2 MB limit", nil))
		return
	}

	contentType := header.Header.Get("Content-Type")
	if _, ok := allowedMediaTypes[contentType]; !ok {
		httputil.RespondError(w, domain.BadRequest("unsupported file type; use jpeg, png, or webp", nil))
		return
	}

	url, err := h.storage.PutObject(r.Context(), prefix, file, header.Size, contentType)
	if err != nil {
		httputil.RespondError(w, domain.Internal(err))
		return
	}

	httputil.Respond(w, http.StatusOK, map[string]string{"url": url})
}
