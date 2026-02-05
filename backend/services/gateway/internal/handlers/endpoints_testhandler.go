package handlers

import (
	"fmt"
	"net/http"
	ct "social-network/shared/go/ct"
	utils "social-network/shared/go/http-utils"
	tele "social-network/shared/go/telemetry"
)

func (h *Handlers) testHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tele.Info(ctx, "test handler called"+fmt.Sprint(r.PathValue("yo")))

		err := utils.WriteJSON(ctx, w, http.StatusOK, map[string]string{
			"message": "this request id is: " + r.Context().Value(ct.ReqID).(string),
		})

		if err != nil {
			tele.Warn(ctx, "failed to send test ACK. @1", "error", err.Error())
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "failed to send logout ACK")
			return
		}
	}
}
