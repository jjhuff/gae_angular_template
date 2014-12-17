package auth

import (
	"crypto/rand"
	"errors"
	"net/http"
	"sync"
	"time"

	"appengine"
	"appengine/datastore"

	"code.google.com/p/go.crypto/bcrypt"
	"github.com/dgrijalva/jwt-go"

	"example.com/user"
)

const (
	cookieName = "auth_jwt"

	keyEntityKind = "SigningKey"
	keyLength     = 32

	authCookieKind   = "ac"
	authCookieAge    = 14 * 24 * time.Hour // 2 weeks
	lostPasswordKind = "lp"
	lostPasswordAge  = 24 * time.Hour
)

var (
	ErrNoCookie      = errors.New("No Auth Cookie")
	ErrInvalidCookie = errors.New("Invalid Auth Cookie")
)

type key struct {
	Key []byte
}

var signingKey struct {
	key
	sync.Once
}

func loadSigningKey(ctx appengine.Context) {
	signingKey.Do(func() {
		ctx.Infof("Loading signing key")
		dsKey := datastore.NewKey(ctx, keyEntityKind, "key", 0, nil)
		err := datastore.RunInTransaction(ctx, func(c appengine.Context) error {
			e := datastore.Get(c, dsKey, &signingKey.key)
			if e == datastore.ErrNoSuchEntity {
				ctx.Infof("Creating new signing key")
				signingKey.Key = make([]byte, keyLength)
				_, e = rand.Read(signingKey.Key)
				if e != nil {
					return e
				}

				_, e = datastore.Put(c, dsKey, &signingKey.key)
				if e != nil {
					return e
				}
				e = nil
			}
			return e
		}, nil)
		if err != nil {
			ctx.Errorf("Failed to created key: %v", err)
		}
	})
	if len(signingKey.Key) == 0 {
		panic("Unable to load signing key")
	}
}

// Create a signed userId
// Format: userId|timestamp|keyName|Signature
func getSignedUserId(ctx appengine.Context, userId user.UserId, kind string, age time.Duration) (string, error) {
	loadSigningKey(ctx)

	// Create the token
	token := jwt.New(jwt.GetSigningMethod("HS256"))
	token.Claims["kind"] = kind
	token.Claims["uid"] = userId.String()
	token.Claims["exp"] = time.Now().Add(age).Unix()

	// Sign and get the complete encoded token as a string
	return token.SignedString(signingKey.Key)
}

func verifySignedUserId(ctx appengine.Context, val, kind string) (user.UserId, error) {
	loadSigningKey(ctx)

	token, err := jwt.Parse(val, func(token *jwt.Token) (interface{}, error) {
		return signingKey.Key, nil
	})
	if err != nil {
		return user.InvalidUser, ErrInvalidCookie
	}

	if token.Claims["kind"] != kind {
		return user.InvalidUser, ErrInvalidCookie
	}

	if userId, ok := token.Claims["uid"].(string); !ok {
		return user.InvalidUser, ErrInvalidCookie
	} else {
		return user.ParseUserId(userId)
	}
}

func SetCookie(ctx appengine.Context, w http.ResponseWriter, u user.User) error {
	cookie, err := getSignedUserId(ctx, u.Id, authCookieKind, authCookieAge)
	if err != nil {
		return ErrNoCookie
	}
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    cookie,
		Path:     "/",
		MaxAge:   int(authCookieAge.Seconds()),
		HttpOnly: true,
		Secure:   !appengine.IsDevAppServer(),
	})
	return nil
}

func ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   cookieName,
		Path:   "/",
		MaxAge: -1,
	})
}

func VerifyCookie(ctx appengine.Context, req *http.Request) (user.UserId, error) {
	cookie, err := req.Cookie(cookieName)
	if err != nil {
		return 0, err
	}

	return verifySignedUserId(ctx, cookie.Value, authCookieKind)
}

func GetLostPasswordToken(ctx appengine.Context, u user.User) (string, error) {
	return getSignedUserId(ctx, u.Id, lostPasswordKind, lostPasswordAge)
}

func VerifyLostPasswordToken(ctx appengine.Context, val string) (user.UserId, error) {
	return verifySignedUserId(ctx, val, lostPasswordKind)
}

// Password Hashing
func HashPassword(pw string) ([]byte, error) {
	const bcryptCost = 13
	return bcrypt.GenerateFromPassword([]byte(pw), bcryptCost)
}

func CheckPassword(pw string, hash []byte) error {
	return bcrypt.CompareHashAndPassword(hash, []byte(pw))
}
