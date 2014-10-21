package example

import (
	"appengine"
	"appengine/mail"
	"bytes"
	"example.com/auth"
	"example.com/user"
	"github.com/ant0ine/go-json-rest/rest"
	"html/template"
	"net/http"
)

var lostPasswordTmpl = template.Must(template.ParseFiles("emails/lost_password.txt"))

type lostPasswordArgs struct {
	ResetURL string
}

func (app *AppContext) RegisterUser(w rest.ResponseWriter, req *rest.Request) {
	ctx := appengine.NewContext(req.Request)
	userDb := user.NewAppEngineUserDB(ctx)

	u := new(struct {
		user.User
		Password string
	})
	err := req.DecodeJsonPayload(&u)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Hash the password
	u.PasswordHash, err = auth.HashPassword(u.Password)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create the user
	err = userDb.Create(&u.User)
	if err != nil {
		if err == user.ErrEmailExists {
			rest.Error(w, err.Error(), http.StatusConflict)
		} else {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Set the login cookie
	err = auth.SetCookie(ctx, w.(http.ResponseWriter), u.User)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteJson(&u.User)
}

func (app *AppContext) LoginUser(w rest.ResponseWriter, req *rest.Request) {
	ctx := appengine.NewContext(req.Request)
	userDb := user.NewAppEngineUserDB(ctx)

	info := new(struct {
		Email    string
		Password string
	})

	err := req.DecodeJsonPayload(&info)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Lookup the user
	u, err := userDb.GetByEmail(info.Email)
	if err == user.ErrUnknownUser {
		rest.Error(w, "Forbidden", http.StatusForbidden)
		return
	} else if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Check password
	err = auth.CheckPassword(info.Password, u.PasswordHash)
	if err != nil {
		rest.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Set the login cookie
	err = auth.SetCookie(ctx, w.(http.ResponseWriter), *u)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteJson(&u)
}

func (app *AppContext) LogoutUser(w rest.ResponseWriter, req *rest.Request) {
	auth.ClearCookie(w.(http.ResponseWriter))
}

func (app *AppContext) ChangePassword(w rest.ResponseWriter, req *rest.Request) {
	ctx := appengine.NewContext(req.Request)
	userDb := user.NewAppEngineUserDB(ctx)

	info := new(struct {
		Old string
		New string
	})

	err := req.DecodeJsonPayload(&info)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check the login cookie
	userId, err := auth.VerifyCookie(ctx, req.Request)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Lookup the user
	u, err := userDb.GetById(userId)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Check password
	err = auth.CheckPassword(info.Old, u.PasswordHash)
	if err != nil {
		rest.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Hash the password
	u.PasswordHash, err = auth.HashPassword(info.New)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Run the update
	err = userDb.Update(u)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteJson(&u)
}

func (app *AppContext) ResetPassword(w rest.ResponseWriter, req *rest.Request) {
	ctx := appengine.NewContext(req.Request)
	userDb := user.NewAppEngineUserDB(ctx)

	info := new(struct {
		Token string
		New   string
	})

	err := req.DecodeJsonPayload(&info)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check the token
	userId, err := auth.VerifyLostPasswordToken(ctx, info.Token)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// Lookup the user
	u, err := userDb.GetById(userId)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Hash the password
	u.PasswordHash, err = auth.HashPassword(info.New)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Run the update
	err = userDb.Update(u)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteJson(&u)
}
func (app *AppContext) SendForgotPassword(w rest.ResponseWriter, req *rest.Request) {
	ctx := appengine.NewContext(req.Request)
	userDb := user.NewAppEngineUserDB(ctx)

	info := new(struct {
		Email string
	})

	err := req.DecodeJsonPayload(&info)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Lookup the user
	u, err := userDb.GetByEmail(info.Email)
	if err == user.ErrUnknownUser {
		rest.Error(w, "Forbidden", http.StatusForbidden)
		return
	} else if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get a password reset token
	resetToken, err := auth.GetLostPasswordToken(ctx, *u)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Send the email
	data := &lostPasswordArgs{
		ResetURL: "http://www.example.com/reset_password/" + resetToken,
	}

	var b bytes.Buffer
	if err := lostPasswordTmpl.Execute(&b, data); err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	msg := &mail.Message{
		Sender:  "example.com <noreply@" + appengine.AppID(ctx) + ".appspotmail.com>",
		To:      []string{info.Email},
		Subject: "Password Reset",
		Body:    b.String(),
	}
	if err := mail.Send(ctx, msg); err != nil {
		ctx.Errorf("Couldn't send email: %v", err)
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if appengine.IsDevAppServer() {
		ctx.Debugf("Reset URL: %s", data.ResetURL)
	}

	ret := new(struct{})
	w.WriteJson(ret)
}

func (app *AppContext) GetUser(w rest.ResponseWriter, req *rest.Request) {
	ctx := appengine.NewContext(req.Request)
	userDb := user.NewAppEngineUserDB(ctx)

	// Check the login cookie
	userId, err := auth.VerifyCookie(ctx, req.Request)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Pase the requested userid
	reqUserId, err := user.ParseUserId(req.PathParam("id"))
	if err != nil {
		rest.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// Verify that they are logged in as the user they are requesting
	// TODO: add support for admin users
	if reqUserId != userId {
		rest.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// Lookup the user
	u, err := userDb.GetById(userId)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteJson(&u)
}

func (app *AppContext) PutUser(w rest.ResponseWriter, req *rest.Request) {
	ctx := appengine.NewContext(req.Request)
	userDb := user.NewAppEngineUserDB(ctx)

	// Check the login cookie
	userId, err := auth.VerifyCookie(ctx, req.Request)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Pase the requested userid
	reqUserId, err := user.ParseUserId(req.PathParam("id"))
	if err != nil {
		rest.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// Verify that they are logged in as the user they are requesting
	// TODO: add support for admin users
	if reqUserId != userId {
		rest.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	// Parse the user from the request
	var u user.User
	err = req.DecodeJsonPayload(&u)
	if err != nil {
		rest.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//TODO: validate the user before saving

	// Run the update
	err = userDb.Update(&u)
	if err != nil {
		if err == user.ErrEmailExists {
			rest.Error(w, err.Error(), http.StatusConflict)
		} else {
			rest.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteJson(&u)
}
