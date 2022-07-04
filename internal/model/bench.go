package model

// SPDX-License-Identifier: EUPL-1.2

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
)

// Bench contains a number of modules with their configuration
type Bench struct {
	gorm.Model
	ID          uuid.UUID     `gorm:"type:uuid;primarykey;default:uuid_generate_v4()" form:"id"`
	UserID      uuid.UUID     `gorm:"type:uuid" form:"-"`
	ShortCode   string        `form:"short_code"`
	User        User          `form:"-"`
	Active      bool          `form:"active"` // i.e. only show current active bench
	Public      bool          `form:"public"`
	CreatedAt   time.Time     `form:"-"`
	UpdatedAt   time.Time     `form:"-"`
	Modules     []BenchModule `form:"-"`
	Name        string        `form:"name,required"`
	Description string        `form:"description"`
}

func (b *Bench) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("bench_uuid", b.ID.String())

	return nil
}

// BeforeUpdate checks if the current user is allowed to do that
func (b *Bench) BeforeUpdate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context.(*gin.Context)

	var tb Bench
	result := tx.First(&tb, b.ID)
	if result.Error != nil {
		return err
	}

	return isAuthorized(ctx, tb.UserID)
}
