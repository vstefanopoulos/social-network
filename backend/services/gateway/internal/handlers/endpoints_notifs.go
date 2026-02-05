package handlers

import (
	"errors"
	"net/http"
	"social-network/shared/gen-go/notifications"
	ce "social-network/shared/go/commonerrors"
	ct "social-network/shared/go/ct"
	"social-network/shared/go/gorpc"
	utils "social-network/shared/go/http-utils"
	"social-network/shared/go/jwt"
	"social-network/shared/go/mapping"
	tele "social-network/shared/go/telemetry"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (h *Handlers) GetUserNotifications() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			tele.Error(ctx, "problem fetching claims")
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "can't find claims")
			return
		}

		v := r.URL.Query()
		limit, err1 := utils.ParamGet(v, "limit", 100, true)
		offset, err2 := utils.ParamGet(v, "offset", 0, true)
		unreadOnly, err3 := utils.ParamGet(v, "read_only", false, true)
		userId := claims.UserId

		notifTypes := []notifications.NotificationType{}
		if err := errors.Join(err1, err2, err3); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		grpcResponse, err := h.NotifService.GetUserNotifications(ctx, &notifications.GetUserNotificationsRequest{
			UserId:     userId,
			Limit:      int32(limit),
			Offset:     int32(offset),
			Types:      notifTypes,
			UnreadOnly: unreadOnly,
		})

		if err != nil {
			err = ce.DecodeProto(err)
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, err.Error())
			return
		}

		httpCode, _ := gorpc.Classify(err)
		err = utils.WriteJSON(ctx, w, httpCode, mapping.PbToNotifications(grpcResponse.Notifications))
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, err.Error())
		}
	}
}

func (s *Handlers) GetUnreadNotificationsCount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			tele.Error(ctx, "problem fetching claims")
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "can't find claims")
			return
		}

		grpcResponse, err := s.NotifService.GetUnreadNotificationsCount(
			ctx,
			&wrapperspb.Int64Value{Value: claims.UserId},
		)
		if err != nil {
			err = ce.DecodeProto(err)
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, err.Error())
			return
		}

		httpCode, _ := gorpc.Classify(err)
		if err := utils.WriteJSON(ctx, w, httpCode, grpcResponse); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, err.Error())
		}
	}
}

func (s *Handlers) MarkNotificationAsRead() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			tele.Error(ctx, "problem fetching claims")
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "can't find claims")
			return
		}

		v := r.URL.Query()
		notifIdRaw, err := utils.ParamGet(v, "notification_id", "", true)

		notifId, err := ct.DecryptId(notifIdRaw)
		if err != nil {
			tele.Warn(ctx, "problem decrypting notification id: @1", "error", err.Error())
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad notification id")
			return
		}

		_, err = s.NotifService.MarkNotificationAsRead(ctx, &notifications.MarkNotificationAsReadRequest{
			NotificationId: notifId.Int64(),
			UserId:         claims.UserId,
		})

		if err != nil {
			err = ce.DecodeProto(err)
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, err.Error())
			return
		}

		if err := utils.WriteJSON(ctx, w, http.StatusOK, nil); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, err.Error())
		}
	}
}

func (s *Handlers) MarkAllAsRead() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			tele.Error(ctx, "problem fetching claims")
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "can't find claims")
			return
		}

		_, err := s.NotifService.MarkAllAsRead(ctx, &wrapperspb.Int64Value{
			Value: claims.UserId,
		})
		if err != nil {
			err = ce.DecodeProto(err)
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, err.Error())
			return
		}

		if err := utils.WriteJSON(ctx, w, http.StatusOK, nil); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, err.Error())
		}
	}
}

func (s *Handlers) DeleteNotification() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			tele.Error(ctx, "problem fetching claims")
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "can't find claims")
			return
		}

		notifId, err := utils.PathValueGet(r, "notification_id", ct.Id(0), true)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad notification id")
			return
		}

		_, err = s.NotifService.DeleteNotification(ctx, &notifications.DeleteNotificationRequest{
			NotificationId: notifId.Int64(),
			UserId:         claims.UserId,
		})
		if err != nil {
			err = ce.DecodeProto(err)
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, err.Error())
			return
		}

		if err := utils.WriteJSON(ctx, w, http.StatusOK, nil); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, err.Error())
		}
	}
}

func (s *Handlers) GetNotificationPreferences() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			tele.Error(ctx, "problem fetching claims")
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "can't find claims")
			return
		}

		grpcResponse, err := s.NotifService.GetNotificationPreferences(
			ctx,
			&wrapperspb.Int64Value{Value: claims.UserId},
		)
		if err != nil {
			err = ce.DecodeProto(err)
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, err.Error())
			return
		}

		httpCode, _ := gorpc.Classify(err)
		if err := utils.WriteJSON(ctx, w, httpCode, grpcResponse); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, err.Error())
		}

		if err := utils.WriteJSON(ctx, w, http.StatusOK, grpcResponse); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, err.Error())
		}
	}
}
