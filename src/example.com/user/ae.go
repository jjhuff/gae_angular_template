package user

import (
	"appengine"
	"appengine/datastore"
	"strings"
)

const (
	UserEntityKind  = "User"
	emailEntityKind = "Email"
)

type Email struct {
}

type AppEngineUserDB struct {
	ctx appengine.Context
}

func NewAppEngineUserDB(c appengine.Context) *AppEngineUserDB {
	return &AppEngineUserDB{
		ctx: c,
	}
}

func (db *AppEngineUserDB) Create(u *User) error {
	var err error

	u.Email = strings.ToLower(u.Email)

	//TODO:
	// * verify password and email validity

	// Check if the user already exists
	err = db.createEmail(u.Email)
	if err != nil {
		return err
	}

	key := datastore.NewIncompleteKey(db.ctx, UserEntityKind, nil)
	key, err = datastore.Put(db.ctx, key, u)
	if err != nil {
		return err
	}

	// Set the new user id
	u.Id = UserId(key.IntID())

	return nil
}

func (db *AppEngineUserDB) GetByEmail(email string) (*User, error) {
	email = strings.ToLower(email)

	u := new(User)
	q := datastore.NewQuery(UserEntityKind).Filter("Email =", email).Limit(1)
	key, err := q.Run(db.ctx).Next(u)
	if err != nil {
		if err == datastore.Done {
			return nil, ErrUnknownUser
		} else {
			return nil, err
		}
	}
	u.Id = UserId(key.IntID())

	return u, nil
}

func (db *AppEngineUserDB) GetById(id UserId) (*User, error) {
	key := datastore.NewKey(db.ctx, UserEntityKind, "", int64(id), nil)

	u := new(User)
	err := datastore.Get(db.ctx, key, u)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil, ErrUnknownUser
		} else {
			return nil, err
		}
	}
	u.Id = UserId(key.IntID())
	return u, nil
}

func (db *AppEngineUserDB) Update(u *User) error {
	// TODO: make sure that the user exists
	key := datastore.NewKey(db.ctx, UserEntityKind, "", int64(u.Id), nil)

	var existingUser User
	err := datastore.Get(db.ctx, key, &existingUser)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return ErrUnknownUser
		} else {
			return err
		}
	}

	// Look for email updates, and do the right thing
	if u.Email != existingUser.Email {
		err = db.updateEmail(existingUser.Email, u.Email)
		if err != nil {
			return err
		}
	}

	// If we aren't setting a new password, make sure we keep the old one
	if len(u.PasswordHash) == 0 {
		u.PasswordHash = existingUser.PasswordHash
	}
	_, err = datastore.Put(db.ctx, key, u)
	return err
}

// Create a new email record if and only if one doesn't exist
func (db *AppEngineUserDB) createEmail(email string) error {
	return datastore.RunInTransaction(db.ctx, func(ctx appengine.Context) error {
		var e Email
		key := datastore.NewKey(ctx, emailEntityKind, email, 0, nil)
		err := datastore.Get(ctx, key, &e)
		if err == nil {
			return ErrEmailExists
		} else if err != datastore.ErrNoSuchEntity {
			return err
		}

		_, err = datastore.Put(ctx, key, &e)
		return err
	}, nil)
}

func (db *AppEngineUserDB) updateEmail(from, to string) error {
	return datastore.RunInTransaction(db.ctx, func(ctx appengine.Context) error {
		var e Email
		toKey := datastore.NewKey(ctx, emailEntityKind, to, 0, nil)
		fromKey := datastore.NewKey(ctx, emailEntityKind, from, 0, nil)

		err := datastore.Get(ctx, toKey, &e)
		if err == nil {
			return ErrEmailExists
		} else if err != datastore.ErrNoSuchEntity {
			return err
		}

		_, err = datastore.Put(ctx, toKey, &e)
		if err != nil {
			return err
		}

		_ = datastore.Delete(ctx, fromKey)
		return nil
	}, &datastore.TransactionOptions{XG: true})
}
