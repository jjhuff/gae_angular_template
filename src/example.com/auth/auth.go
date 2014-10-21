package auth

import (
	"appengine"
	"bytes"
	"code.google.com/p/go.crypto/bcrypt"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"example.com/user"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	cookieName   = "auth"
	certCacheAge = 1 * time.Hour

	authCookieKind   = "ac"
	authCookieAge    = 14 * 24 * time.Hour // 2 weeks
	lostPasswordKind = "lp"
	lostPasswordAge  = 24 * time.Hour
)

var (
	ErrNoCookie      = errors.New("No Auth Cookie")
	ErrInvalidCookie = errors.New("Invalid Auth Cookie")
)

var certCache struct {
	certs   map[string]*x509.Certificate
	expires time.Time
	mutex   sync.RWMutex
}

func updateCertCache(ctx appengine.Context) error {
	// Update cache
	ctx.Infof("Fetch certs")
	certs, err := appengine.PublicCertificates(ctx)
	if err != nil {
		return err
	}

	// Parse and map
	certCache.certs = map[string]*x509.Certificate{}
	for _, cert := range certs {
		block, _ := pem.Decode(cert.Data)
		x509cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return errors.New("Unable to parse public certificate")
		}
		certCache.certs[cert.KeyName] = x509cert
	}
	certCache.expires = time.Now().Add(certCacheAge)
	return nil
}

func findCertificate(ctx appengine.Context, keyName string) *x509.Certificate {
	// Check if we have a valid cache
	certCache.mutex.RLock()
	if certCache.expires.After(time.Now()) {
		cert := certCache.certs[keyName]
		certCache.mutex.RUnlock()
		return cert
	}
	certCache.mutex.RUnlock()

	certCache.mutex.Lock()
	defer certCache.mutex.Unlock()
	err := updateCertCache(ctx)
	if err != nil {
		ctx.Errorf("Error updating certificate cache: %v", err)
		return nil
	}

	return certCache.certs[keyName]

}

// Create a signed userId
// Format: userId|timestamp|keyName|Signature
func getSignedUserId(ctx appengine.Context, userId user.UserId, kind string) (string, error) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	const sep = "|"
	val := kind + sep + userId.String() + sep + timestamp

	var key string
	var sig []byte
	var err error

	// v1.8.8 of the SDK has a broken SignBytes
	if !appengine.IsDevAppServer() {
		key, sig, err = appengine.SignBytes(ctx, []byte(val))
		if err != nil {
			return "", err
		}
	} else {
		sig = []byte(val + "_abc")
		certs, _ := appengine.PublicCertificates(ctx)
		key = certs[0].KeyName
	}

	ret := val + sep + key + sep + base64.URLEncoding.EncodeToString(sig)
	return ret, nil
}

func verifySignedUserId(ctx appengine.Context, val, kind string, age time.Duration) (user.UserId, error) {
	// Parse the value
	valSplit := strings.Split(val, "|")
	if len(valSplit) != 5 {
		return 0, ErrInvalidCookie
	}
	userId, err := user.ParseUserId(valSplit[1])
	if err != nil {
		return 0, ErrInvalidCookie
	}
	timestampInt, err := strconv.ParseInt(valSplit[2], 10, 64)
	if err != nil {
		return 0, ErrInvalidCookie
	}
	timestamp := time.Unix(timestampInt, 0)
	expires := time.Now().Add(age)
	if timestamp.After(expires) {
		if err != nil {
			return 0, ErrInvalidCookie
		}
	}

	// Locate the certificate
	cert := findCertificate(ctx, valSplit[3])
	if cert == nil {
		return 0, ErrInvalidCookie
	}

	// Decode the base64 signature
	sig, err := base64.URLEncoding.DecodeString(valSplit[4])
	if err != nil {
		return 0, ErrInvalidCookie
	}

	sigVal := strings.Join(valSplit[0:3], "|")
	if !appengine.IsDevAppServer() {
		err = cert.CheckSignature(x509.SHA256WithRSA, []byte(sigVal), sig)
		if err != nil {
			return 0, ErrInvalidCookie
		}
		// check sig
	} else {
		if !bytes.Equal(sig, []byte(sigVal+"_abc")) {
			return 0, ErrInvalidCookie
		}
	}

	return userId, nil
}

func SetCookie(ctx appengine.Context, w http.ResponseWriter, user user.User) error {
	cookie, err := getSignedUserId(ctx, user.Id, authCookieKind)
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

	return verifySignedUserId(ctx, cookie.Value, authCookieKind, authCookieAge)
}

// Password Hashing
func HashPassword(pw string) ([]byte, error) {
	const bcryptCost = 13
	return bcrypt.GenerateFromPassword([]byte(pw), bcryptCost)
}

func CheckPassword(pw string, hash []byte) error {
	return bcrypt.CompareHashAndPassword(hash, []byte(pw))
}

func GetLostPasswordToken(ctx appengine.Context, user user.User) (string, error) {
	return getSignedUserId(ctx, user.Id, lostPasswordKind)
}

func VerifyLostPasswordToken(ctx appengine.Context, val string) (user.UserId, error) {
	return verifySignedUserId(ctx, val, lostPasswordKind, lostPasswordAge)
}
