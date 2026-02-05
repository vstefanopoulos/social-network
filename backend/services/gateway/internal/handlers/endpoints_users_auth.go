package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"social-network/shared/gen-go/media"
	"social-network/shared/gen-go/users"
	ct "social-network/shared/go/ct"
	utils "social-network/shared/go/http-utils"
	"social-network/shared/go/jwt"
	"social-network/shared/go/models"
	tele "social-network/shared/go/telemetry"
	"time"
)

func (h *Handlers) loginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tele.Info(ctx, "login handler called")

		//READ REQUEST BODY
		type loginHttpRequest struct {
			Identifier ct.Identifier `json:"identifier"`
			Password   ct.Password   `json:"password"`
		}

		httpReq := loginHttpRequest{}

		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()
		err := decoder.Decode(&httpReq)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, err.Error())
			return
		}

		httpReq.Password, err = httpReq.Password.Hash()
		if err != nil {
			tele.Error(ctx, "failed to hash password. @1", "error", err.Error())
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "could not hash password")
			return
		}

		//VALIDATE INPUT
		if err := ct.ValidateStruct(httpReq); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, err.Error())
			return
		}
		tele.Debug(ctx, "login password:", "password", httpReq.Password.String())
		//MAKE GRPC REQUEST
		gRpcReq := users.LoginRequest{
			Identifier: httpReq.Identifier.String(),
			Password:   httpReq.Password.String(),
		}

		tele.Debug(ctx, "login password request:", "password", httpReq.Password.String())

		resp, err := h.UsersService.LoginUser(r.Context(), &gRpcReq)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, err.Error())
			return
		}

		//PREPARE SUCCESS RESPONSE
		now := time.Now().Unix()
		exp := time.Now().AddDate(0, 6, 0).Unix() // six months from now

		claims := jwt.Claims{
			UserId: resp.UserId,
			Iat:    now,
			Exp:    exp,
		}

		token, err := jwt.CreateToken(claims)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "token generation failed")
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "jwt",
			Value:    token,
			Path:     "/",
			Expires:  time.Unix(exp, 0),
			HttpOnly: true,
			Secure:   false, //TODO: set to true in production
			SameSite: http.SameSiteLaxMode,
		})

		httpResp := models.User{
			UserId:    ct.Id(resp.UserId),
			Username:  ct.Username(resp.Username),
			AvatarId:  ct.Id(resp.Avatar),
			AvatarURL: resp.AvatarUrl,
		}

		//SEND RESPONSE
		err = utils.WriteJSON(ctx, w, http.StatusCreated, httpResp)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusUnauthorized, "failed to send login ACK")
			return
		}
	}
}

func (h *Handlers) registerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tele.Info(ctx, "register handler called")

		// Check if user already logged in
		cookie, _ := r.Cookie("jwt")
		if cookie != nil {
			_, err := jwt.ParseAndValidate(cookie.Value)
			if err == nil {
				utils.ErrorJSON(ctx, w, http.StatusForbidden, "Already logged in. Log out to register.")
				return
			}
		}

		//READ REQUEST BODY
		type registerHttpRequest struct {
			Username    ct.Username    `json:"username,omitempty" validate:"nullable"`
			FirstName   ct.Name        `json:"first_name,omitempty"`
			LastName    ct.Name        `json:"last_name,omitempty"`
			DateOfBirth ct.DateOfBirth `json:"date_of_birth,omitempty"`
			About       ct.About       `json:"about,omitempty" validate:"nullable"`
			Public      bool           `json:"public,omitempty"`
			Email       ct.Email       `json:"email,omitempty"`
			Password    ct.Password    `json:"password,omitempty"`

			AvatarName string `json:"avatar_name"`
			AvatarSize int64  `json:"avatar_size"`
			AvatarType string `json:"avatar_type"`
		}

		httpReq := registerHttpRequest{}

		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()
		if err := decoder.Decode(&httpReq); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, err.Error())
			return
		}

		tele.Debug(ctx, "register request contents: "+fmt.Sprint(httpReq))

		hashedPass, err := httpReq.Password.Hash()
		if err != nil {
			tele.Error(ctx, "failed to hash password. @1", "error", err.Error())
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "could not hash password")
			return
		}

		if err := ct.ValidateStruct(httpReq); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, err.Error())
			return
		}

		var AvatarId ct.Id
		var uploadURL string
		if httpReq.AvatarSize != 0 {
			exp := time.Duration(10 * time.Minute).Seconds()
			mediaRes, err := h.MediaService.UploadImage(r.Context(), &media.UploadImageRequest{
				Filename:          httpReq.AvatarName,
				MimeType:          httpReq.AvatarType,
				SizeBytes:         httpReq.AvatarSize,
				Visibility:        media.FileVisibility_PUBLIC,
				Variants:          []media.FileVariant{media.FileVariant_THUMBNAIL},
				ExpirationSeconds: int64(exp),
			})
			if err != nil {
				utils.ErrorJSON(ctx, w, http.StatusInternalServerError, err.Error())
				return
			}
			AvatarId = ct.Id(mediaRes.FileId)
			uploadURL = mediaRes.GetUploadUrl()
		}

		//MAKE GRPC REQUEST
		gRpcReq := users.RegisterUserRequest{
			Username:    string(httpReq.Username),
			FirstName:   string(httpReq.FirstName),
			LastName:    string(httpReq.LastName),
			DateOfBirth: httpReq.DateOfBirth.ToProto(),
			About:       string(httpReq.About),
			Public:      httpReq.Public,
			Email:       string(httpReq.Email),
			Password:    string(hashedPass),
			Avatar:      AvatarId.Int64(),
		}

		resp, err := h.UsersService.RegisterUser(r.Context(), &gRpcReq)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, err.Error())
			return
		}

		//PREPARE SUCCESS RESPONSE
		now := time.Now().Unix()
		exp := time.Now().AddDate(0, 6, 0).Unix() // six months from now

		claims := jwt.Claims{
			UserId: resp.UserId,
			Iat:    now,
			Exp:    exp,
		}

		token, err := jwt.CreateToken(claims)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "token generation failed")
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "jwt",
			Value:    token,
			Path:     "/",
			Expires:  time.Unix(exp, 0),
			HttpOnly: true,
			Secure:   false, //TODO: set to true in production
			SameSite: http.SameSiteLaxMode,
		})

		type httpResponse struct {
			UserId    ct.Id
			Username  ct.Username
			FileId    ct.Id
			UploadUrl string
		}
		httpResp := httpResponse{
			UserId:    ct.Id(resp.UserId),
			Username:  ct.Username(resp.Username),
			FileId:    AvatarId,
			UploadUrl: uploadURL}

		//SEND RESPONSE
		if err := utils.WriteJSON(ctx, w, http.StatusCreated, httpResp); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "failed to send registration ACK")
			return
		}
	}
}

func (h *Handlers) logoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tele.Info(ctx, "logout handler called")
		//CLEAR COOKIE
		http.SetCookie(w, &http.Cookie{
			Name:     "jwt",
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			Secure:   false, //TODO: set to true in production
			SameSite: http.SameSiteLaxMode,
		})

		//SEND RESPONSE
		if err := utils.WriteJSON(ctx, w, http.StatusOK, "logged out successfully"); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "failed to send logout ACK")
			return
		}
	}
}

// Returns status ok if passed Auth
func (h *Handlers) authStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.WriteJSON(r.Context(), w, http.StatusOK, "user is logged in")
	}
}

func (s *Handlers) updateUserEmail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		type reqBody struct {
			Email string `json:"email"`
		}

		body, err := utils.JSON2Struct(&reqBody{}, r)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "Bad JSON data received")
			return
		}

		req := &users.UpdateEmailRequest{
			UserId: claims.UserId,
			Email:  body.Email,
		}

		_, err = s.UsersService.UpdateUserEmail(ctx, req)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not update email: "+err.Error())
			return
		}

		utils.WriteJSON(ctx, w, http.StatusOK, nil)
	}
}

// TODO should probably be done using a specific link / needs extra validation
func (s *Handlers) updateUserPassword() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		type reqBody struct {
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}

		body, err := utils.JSON2Struct(&reqBody{}, r)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "Bad JSON data received")
			return
		}
		_ = body

		tele.Info(ctx, "updating passwords. @1 @2", "old", body.OldPassword, "new", body.NewPassword)

		oldPassword, err := ct.Password(body.OldPassword).Hash()
		if err != nil {
			tele.Error(ctx, "failed to hash password @1", "error", err.Error())
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "could not hash password")
			return
		}

		newPassword, err := ct.Password(body.NewPassword).Hash()
		if err != nil {
			tele.Error(ctx, "failed to hash password @1", "error", err.Error())
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "could not hash password")
			return
		}

		tele.Info(ctx, "hashed passwords. @1 @2", "old", oldPassword.String(), "new", newPassword.String())

		req := &users.UpdatePasswordRequest{
			UserId:      claims.UserId,
			OldPassword: oldPassword.String(),
			NewPassword: newPassword.String(),
		}

		_, err = s.UsersService.UpdateUserPassword(ctx, req)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not update password: "+err.Error())
			return
		}

		utils.WriteJSON(ctx, w, http.StatusOK, nil)
	}
}
