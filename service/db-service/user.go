package dbService

import (
	"SecKill/dao"
	"SecKill/model"
)

func GetUser(userName string) (model.User, error) {
	user := model.User{}
	operation := dao.Db.Where("username = ?", userName).First(&user)
	return user, operation.Error
}
