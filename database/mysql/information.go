package mysql

import (
	"spider/types"
	"github.com/pkg/errors"
	"time"
)

func (m *mysql) Add(info types.Information) (id uint64, err error) {
	info.CreateTime = time.Now().Unix()
	tx := m.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	if err := tx.Create(&info).Error; err != nil {
		return 0, errors.WithMessage(err, "create information")
	}
	return info.Id, nil
}

func (m *mysql) GetInfoByTitle(title string) (*types.Information, error) {
	var result types.Information
	err := m.db.Where("title = ?", title).First(&result).Error
	return &result, handleErr(err, "information")
}

func (m *mysql) GetInfoById(id uint64) (*types.Information, error) {
	var result types.Information
	err := m.db.Where("id = ?", id).First(&result).Error
	return &result, handleErr(err, "information")
}

func (m *mysql) GetInfoByContext(context string) (*types.Information, error) {
	var result types.Information
	err := m.db.Where("context = ?", context).First(&result).Error
	return &result, handleErr(err, "information")
}

func (m *mysql) UpdateInfoById(information *types.Information) (err error) {
	now := time.Now().Unix()
	tx := m.db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	if err := tx.Model(&information).Update(map[string]interface{}{"context": information.Context, "create_time": now, "title": information.Title}).Error; err != nil {
		return errors.WithMessage(err, "create information")
	}
	return nil
}

func (m *mysql) GetInfoByTitleOrContext(title, context string) (*types.Information, error) {
	var result types.Information
	err := m.db.Where("title = ? or context = ?", title, context).First(&result).Error
	return &result, handleErr(err, "information")
}