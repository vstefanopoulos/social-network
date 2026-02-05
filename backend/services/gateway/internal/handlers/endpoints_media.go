package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"social-network/shared/gen-go/media"
	ct "social-network/shared/go/ct"
	utils "social-network/shared/go/http-utils"
	"social-network/shared/go/mapping"
	tele "social-network/shared/go/telemetry"
)

func (h *Handlers) validateFileUpload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		type validateUploadReq struct {
			FileId    ct.Id
			ReturnURL bool `json:"return_url"`
		}
		httpReq := validateUploadReq{}

		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()
		if err := decoder.Decode(&httpReq); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, err.Error())
			return
		}

		var err error
		httpReq.FileId, err = utils.PathValueGet(r, "file_id", ct.Id(0), true)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		res, err := h.MediaService.ValidateUpload(
			r.Context(),
			&media.ValidateUploadRequest{
				FileId:    httpReq.FileId.Int64(),
				ReturnUrl: httpReq.ReturnURL},
		)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, err.Error())
			return
		}

		tele.Info(ctx, "Gateway: Successfully validated file upload for @1", "fileId", httpReq.FileId)

		type httpResp struct {
			DownloadUrl string `json:"download_url"`
		}

		if err := utils.WriteJSON(ctx, w, http.StatusCreated, &httpResp{DownloadUrl: res.GetDownloadUrl()}); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, fmt.Sprintf("failed to validate file %v", httpReq.FileId))
			return
		}
	}
}

// Only valid for public images
func (h *Handlers) getImageUrl() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		imageId, err1 := utils.PathValueGet(r, "image_id", ct.Id(0), true)
		variant, err2 := utils.PathValueGet(r, "variant", ct.FileVariant("thumb"), false)
		if err := errors.Join(err1, err2); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		res, err := h.MediaService.GetImage(r.Context(), &media.GetImageRequest{
			ImageId: imageId.Int64(),
			Variant: mapping.CtToPbFileVariant(variant),
		})
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, err.Error())
			return
		}

		type httpResp struct {
			DownloadUrl string `json:"download_url"`
		}

		httpRes := &httpResp{DownloadUrl: res.DownloadUrl}
		if err := utils.WriteJSON(ctx, w, http.StatusOK, httpRes); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "failed to send registration ACK")
			return
		}
	}
}
