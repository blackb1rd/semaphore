package api

import (
	"bytes"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/semaphoreui/semaphore/api/helpers"
	"github.com/semaphoreui/semaphore/db"
	log "github.com/sirupsen/logrus"
	"image/png"
	"net/http"

	"github.com/gorilla/context"
	"github.com/semaphoreui/semaphore/util"
)

type minimalUser struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	currentUser := context.Get(r, "user").(*db.User)
	users, err := helpers.Store(r).GetUsers(db.RetrieveQueryParams{
		Filter: r.URL.Query().Get("s"),
	})

	if err != nil {
		panic(err)
	}

	if currentUser.Admin {
		helpers.WriteJSON(w, http.StatusOK, users)
	} else {
		var result = make([]minimalUser, 0)

		for _, user := range users {
			result = append(result, minimalUser{
				ID:       user.ID,
				Name:     user.Name,
				Username: user.Username,
			})
		}

		helpers.WriteJSON(w, http.StatusOK, result)
	}
}

func addUser(w http.ResponseWriter, r *http.Request) {
	var user db.UserWithPwd
	if !helpers.Bind(w, r, &user) {
		return
	}

	editor := context.Get(r, "user").(*db.User)
	if !editor.Admin {
		log.Warn(editor.Username + " is not permitted to create users")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var err error
	var newUser db.User

	if user.External {
		newUser, err = helpers.Store(r).CreateUserWithoutPassword(user.User)
	} else {
		newUser, err = helpers.Store(r).CreateUser(user)
	}

	if err != nil {
		log.Warn(editor.Username + " is not created: " + err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	helpers.WriteJSON(w, http.StatusCreated, newUser)
}
func readonlyUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := helpers.GetIntParam("user_id", w, r)

		if err != nil {
			return
		}

		user, err := helpers.Store(r).GetUser(userID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		editor := context.Get(r, "user").(*db.User)

		if !editor.Admin && editor.ID != user.ID {
			user = db.User{
				ID:       user.ID,
				Username: user.Username,
				Name:     user.Name,
			}
		}

		context.Set(r, "_user", user)
		next.ServeHTTP(w, r)
	})
}

func getUserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := helpers.GetIntParam("user_id", w, r)

		if err != nil {
			return
		}

		user, err := helpers.Store(r).GetUser(userID)

		if err != nil {
			helpers.WriteError(w, err)
			return
		}

		editor := context.Get(r, "user").(*db.User)

		if !editor.Admin && editor.ID != user.ID {
			log.Warn(editor.Username + " is not permitted to edit users")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		context.Set(r, "_user", user)
		next.ServeHTTP(w, r)
	})
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	targetUser := context.Get(r, "_user").(db.User)
	editor := context.Get(r, "user").(*db.User)

	var user db.UserWithPwd
	if !helpers.Bind(w, r, &user) {
		return
	}

	if !editor.Admin && (user.Pro && !targetUser.Pro) {
		log.Warn(editor.Username + " is not permitted to mark users as Pro")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if !editor.Admin && editor.ID != targetUser.ID {
		log.Warn(editor.Username + " is not permitted to edit users")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if editor.ID == targetUser.ID && targetUser.Admin != user.Admin {
		log.Warn("User can't edit his own role")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if targetUser.External && targetUser.Username != user.Username {
		log.Warn("Username is not editable for external users")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user.ID = targetUser.ID
	if err := helpers.Store(r).UpdateUser(user); err != nil {
		log.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func updateUserPassword(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "_user").(db.User)
	editor := context.Get(r, "user").(*db.User)

	var pwd struct {
		Pwd string `json:"password"`
	}

	if !editor.Admin && editor.ID != user.ID {
		log.Warn(editor.Username + " is not permitted to edit users")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if user.External {
		log.Warn("Password is not editable for external users")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !helpers.Bind(w, r, &pwd) {
		return
	}

	if err := helpers.Store(r).SetUserPassword(user.ID, pwd.Pwd); err != nil {
		util.LogWarning(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "_user").(db.User)
	editor := context.Get(r, "user").(*db.User)

	if !editor.Admin && editor.ID != user.ID {
		log.Warn(editor.Username + " is not permitted to delete users")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err := helpers.Store(r).DeleteUser(user.ID); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}

func totpQr(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "_user").(db.User)

	if user.Totp == nil {
		helpers.WriteErrorStatus(w, "TOTP not enabled", http.StatusNotFound)
		return
	}

	key, err := otp.NewKeyFromURL(user.Totp.URL)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	image, err := key.Image(256, 256)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	var buf bytes.Buffer
	err = png.Encode(&buf, image)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}
	pngBytes := buf.Bytes()

	w.Header().Add("Content-Type", "image/png")
	_, err = w.Write(pngBytes)
}

func enableTotp(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "_user").(db.User)

	if !util.Config.Auth.Totp.Enabled {
		helpers.WriteErrorStatus(w, "TOTP not enabled", http.StatusBadRequest)
		return
	}

	if user.Totp != nil {
		helpers.WriteErrorStatus(w, "TOTP already enabled", http.StatusBadRequest)
		return
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Semaphore",
		AccountName: user.Email,
	})

	if err != nil {
		http.Error(w, "Error generating key", http.StatusInternalServerError)
		return
	}

	var code, hash string

	if util.Config.Auth.Totp.AllowRecovery {
		code, hash, err = util.GenerateRecoveryCode()
		if err != nil {
			helpers.WriteError(w, err)
			return
		}
	}

	newTotp, err := helpers.Store(r).AddTotpVerification(user.ID, key.URL(), hash)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	newTotp.RecoveryCode = code

	helpers.WriteJSON(w, http.StatusOK, newTotp)
}

func disableTotp(w http.ResponseWriter, r *http.Request) {
	user := context.Get(r, "_user").(db.User)
	if user.Totp == nil {
		helpers.WriteErrorStatus(w, "TOTP not enabled", http.StatusBadRequest)
		return
	}

	totpID, err := helpers.GetIntParam("totp_id", w, r)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	err = helpers.Store(r).DeleteTotpVerification(user.ID, totpID)
	if err != nil {
		helpers.WriteError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
