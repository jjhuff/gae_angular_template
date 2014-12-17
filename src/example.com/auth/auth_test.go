package auth

import (
	"appengine/aetest"
	"example.com/user"
	"github.com/stretchr/testify/assert"
	"testing"
)

func getAppEngineContext(t *testing.T) aetest.Context {
	c, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestBadCookie(t *testing.T) {
	ctx := getAppEngineContext(t)
	defer ctx.Close()

	_, err := verifySignedUserId(ctx, "uw4rtgdfhg", authCookieKind)
	assert.Equal(t, err, ErrInvalidCookie)
}

func TestCookies(t *testing.T) {
	const testUserId = 349524596
	ctx := getAppEngineContext(t)
	defer ctx.Close()

	cookieVal, err := getSignedUserId(ctx, testUserId, authCookieKind, authCookieAge)
	assert.NoError(t, err)
	assert.NotEmpty(t, cookieVal)

	userId, err := verifySignedUserId(ctx, cookieVal, authCookieKind)
	assert.NoError(t, err)
	assert.Equal(t, userId, testUserId)

	brokenCookieVal := "1" + cookieVal
	userId, err = verifySignedUserId(ctx, brokenCookieVal, authCookieKind)
	assert.Equal(t, err, ErrInvalidCookie)
}

func TestLostPassword(t *testing.T) {
	ctx := getAppEngineContext(t)
	defer ctx.Close()

	var u = user.User{
		Id: 394585,
	}

	val, err := GetLostPasswordToken(ctx, u)
	assert.NoError(t, err)
	assert.NotEmpty(t, val)

	id, err := VerifyLostPasswordToken(ctx, val)
	assert.NoError(t, err)
	assert.Equal(t, id, u.Id)
}
