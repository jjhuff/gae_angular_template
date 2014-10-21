package user

import (
	"appengine/aetest"
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

func TestCreate(t *testing.T) {
	ctx := getAppEngineContext(t)
	defer ctx.Close()
	db := NewAppEngineUserDB(ctx)

	u := &User{
		Email: "foo@example.com",
	}
	err := db.Create(u)
	assert.NoError(t, err)
	assert.NotEqual(t, u.Id, 0)
}

func TestEmailUpdate(t *testing.T) {
	ctx := getAppEngineContext(t)
	defer ctx.Close()
	db := NewAppEngineUserDB(ctx)

	u := User{
		Email:        "u1@example.com",
		PasswordHash: []byte("somebytes"),
	}

	u1 := u
	err := db.Create(&u1)
	assert.NoError(t, err)

	u1.Email = "u2@example.com"
	u1.PasswordHash = nil // assume that updates won't have a password hash
	err = db.Update(&u1)
	assert.NoError(t, err)
	_, err = db.GetById(u1.Id) // force a commit
	assert.NoError(t, err)

	_, err = db.GetByEmail(u.Email)
	assert.Equal(t, err, ErrUnknownUser)

	u2, err := db.GetByEmail(u1.Email)
	assert.NoError(t, err)
	assert.Equal(t, u1.Email, u2.Email)
	assert.Equal(t, u.PasswordHash, u2.PasswordHash)
}

func TestDuplicateEmail_Create(t *testing.T) {
	ctx := getAppEngineContext(t)
	defer ctx.Close()
	db := NewAppEngineUserDB(ctx)

	u1 := User{
		Email: "foo@example.com",
	}
	u2 := u1
	err := db.Create(&u1)
	assert.NoError(t, err)

	err = db.Create(&u2)
	assert.Equal(t, err, ErrEmailExists)
}

func TestDuplicateEmail_Update(t *testing.T) {
	ctx := getAppEngineContext(t)
	defer ctx.Close()
	db := NewAppEngineUserDB(ctx)

	u1 := User{
		Email: "u1@example.com",
	}
	err := db.Create(&u1)
	assert.NoError(t, err)

	u2 := User{
		Email: "u2@example.com",
	}
	err = db.Create(&u2)
	assert.NoError(t, err)

	u1.Email = "collide@example.com"
	err = db.Update(&u1)
	assert.NoError(t, err)

	u2.Email = "collide@example.com"
	err = db.Update(&u2)
	assert.Equal(t, err, ErrEmailExists)
}
