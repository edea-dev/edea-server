package model

import (
	"reflect"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Model interface defines which methods our models need to implement
type Model interface {
	MarshalZerologObject(e *zerolog.Event) // logger method to attach important information to potential log output
	GetMembers() ([]*User, error)          // can return empty list if there's just none
	GetModules() ([]*Module, error)        // can return empty list if there's just none
	Validate(u *User) error                // validate if a user is allowed to make those changes and if it makes sense
}

// DB is the global instance of the database connection
var DB *pg.DB

// CreateTables initially creates the tables in the database
func CreateTables() {
	// register many2many tables
	orm.RegisterTable((*BenchModule)(nil))

	// TODO:
	models := []interface{}{
		(*Bench)(nil),
		(*Profile)(nil),
		(*Module)(nil),
		(*Repository)(nil),
		(*User)(nil),
		(*BenchModule)(nil),
	}

	for _, model := range models {
		err := DB.Model(model).CreateTable(&orm.CreateTableOptions{
			Temp: true,
		})
		if err != nil {
			log.Fatal().Err(err).Msgf("could not create table for %s", reflect.TypeOf(model).String())
		}
	}
}

// updateModel runs update on a generic model
func updateModel(model Model, sub string) error {
	if sub == "" {
		return ErrNoSuchSubject
	}

	user := &User{AuthUUID: sub}

	// get current user info
	if err := DB.Model(user).Column("user.*").Relation("Profile").Select(); err != nil {
		if err == pg.ErrNoRows {
			return ErrUnauthorized
		}
		log.Panic().Err(err).Msgf("could not get user")
	}

	// validate if someone is allowed to make those changes and the changes are valid themselves
	if err := model.Validate(user); err != nil {
		return err
	}

	// check if the user is a member
	members, err := model.GetMembers()
	if err != nil {
		return err
	}
	if len(members) > 0 {
		ok := isAuthorized(members, sub)
		if !ok {
			return ErrUnauthorized
		}
	}

	// check if it contains modules
	modules, err := model.GetModules()
	if err != nil {
		return err
	}

	if len(modules) > 0 {
		// TODO: what do we need to check when a model has modules?
	}

	m := model
	log.Info().Msgf("%+v, %+v", model, m)

	_, err = DB.Model(&m).WherePK().UpdateNotZero()

	if err != nil {
		log.Error().Err(err).Msgf("could not update model: %v", err)
		return err
	}

	// log the action if it was performed by an admin
	if user.IsAdmin {
		log.Info().EmbedObject(model).Str("admin_auth_uuid", sub).Msg("information changed by admin")
	}

	return nil
}

// createModel runs insert on a generic model
func createModel(model Model, sub string) (interface{}, error) {
	if sub == "" {
		return nil, ErrNoSuchSubject
	}

	user := &User{AuthUUID: sub}

	// get current user info
	if err := DB.Model(user).Column("user.*").Relation("Profile").Select(); err != nil {
		if err == pg.ErrNoRows {
			return nil, ErrUnauthorized
		}
		log.Panic().Err(err).Msgf("could not get user")
	}

	// validate if someone is allowed to make those changes and the changes are valid themselves
	if err := model.Validate(user); err != nil {
		return nil, err
	}

	// check if the user is a member
	members, err := model.GetMembers()
	if err != nil {
		return nil, err
	}
	if len(members) > 0 {
		ok := isAuthorized(members, sub)
		if !ok {
			return nil, ErrUnauthorized
		}
	}

	// check if it contains modules
	modules, err := model.GetModules()
	if err != nil {
		return nil, err
	}

	if len(modules) > 0 {
		// TODO: what do we need to check when a model has modules?
	}

	_, err = DB.Model(&model).Insert()

	if err != nil {
		log.Error().Err(err).Msgf("could not update model: %v", err)
		return nil, err
	}

	// log the action if it was performed by an admin
	if user.IsAdmin {
		log.Info().EmbedObject(model).Str("admin_auth_uuid", sub).Msg("information changed by admin")
	}

	return nil, nil
}
