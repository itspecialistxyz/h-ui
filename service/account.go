package service

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"h-ui/dao"
	"h-ui/model/bo"
	"h-ui/model/constant"
	"h-ui/model/dto"
	"h-ui/model/entity"
	"h-ui/model/vo"
	"h-ui/util"
	"strings"

	"github.com/sirupsen/logrus"
)

func Login(username string, pass string) (string, error) {
	account, err := dao.GetAccount("username = ? and role = 'admin' and deleted = 0", username)
	if err != nil {
		// If account not found, return wrong password error for security
		if err.Error() == constant.WrongPassword || strings.Contains(err.Error(), "record not found") {
			return "", errors.New(constant.WrongPassword)
		}
		return "", err // Other DB errors
	}

	if account.Pass == nil || *account.Pass == "" {
		return "", errors.New(constant.WrongPassword) // No password set
	}
	storedHash := *account.Pass
	authenticated := false

	if strings.HasPrefix(storedHash, util.Argon2idPrefix) {
		// Argon2id hash
		match, err := util.VerifyPassword(pass, storedHash)
		if err != nil {
			logrus.Errorf("Error verifying Argon2id password for user %s: %v", username, err)
			return "", errors.New(constant.WrongPassword) // Treat verification error as auth failure
		}
		if match {
			authenticated = true
		}
	} else {
		// Assume old SHA224 hash
		if util.SHA224String(pass) == storedHash {
			authenticated = true
			// Migrate to Argon2id
			newArgon2idHash, err := util.HashPassword(pass)
			if err == nil {
				if errUpdate := dao.UpdateAccount([]int64{*account.Id}, map[string]interface{}{"pass": newArgon2idHash}); errUpdate != nil {
					logrus.Errorf("Failed to migrate password to Argon2id for user %s: %v", username, errUpdate)
					// Do not fail login if migration fails, user is already authenticated
				} else {
					logrus.Infof("Successfully migrated password to Argon2id for user %s", username)
				}
			} else {
				logrus.Errorf("Failed to hash password with Argon2id during migration for user %s: %v", username, err)
				// Do not fail login if hashing fails during migration
			}
		}
	}

	if !authenticated {
		return "", errors.New(constant.WrongPassword)
	}

	accountBo := bo.AccountBo{
		Id:       *account.Id,
		Username: *account.Username,
		Roles:    []string{*account.Role},
		Deleted:  *account.Deleted,
	}
	return GenToken(accountBo)
}

func PageAccount(accountPageDto dto.AccountPageDto) ([]entity.Account, int64, error) {
	return dao.PageAccount(accountPageDto)
}

func SaveAccount(account entity.Account) error {
	// If it's a new account and ConPass is not set, generate a random one.
	if (account.Id == nil || *account.Id == 0) && (account.ConPass == nil || *account.ConPass == "") {
		newRandomConPass, err := util.RandomString(16)
		if err != nil {
			logrus.Errorf("Failed to generate random ConPass for new account %s: %v", *account.Username, err)
			return fmt.Errorf("failed to generate random ConPass: %w", err)
		}
		account.ConPass = &newRandomConPass
		logrus.Infof("Generated random ConPass for new account %s", *account.Username)
	}
	_, err := dao.SaveAccount(account)
	return err
}

func DeleteAccount(ids []int64) error {
	return dao.DeleteAccount(ids)
}

func UpdateAccount(account entity.Account) error {
	updates := map[string]interface{}{}
	if account.Username != nil && *account.Username != "" {
		updates["username"] = *account.Username
	}
	if account.Pass != nil && *account.Pass != "" {
		newPlaintextPassword := *account.Pass
		newArgon2idHash, err := util.HashPassword(newPlaintextPassword)
		if err != nil {
			logrus.Errorf("Failed to hash new password for user %s during update: %v", *account.Username, err)
			return fmt.Errorf("failed to hash new password: %w", err) // Return error if hashing fails
		}
		updates["pass"] = newArgon2idHash
	}
	// If ConPass is explicitly provided in the update, use that value directly.
	// Otherwise, ConPass is not changed.
	if account.ConPass != nil {
		if *account.ConPass == "" { // Allow clearing ConPass by providing an empty string
			updates["con_pass"] = ""
		} else {
			updates["con_pass"] = *account.ConPass
		}
	}
	if account.Quota != nil {
		updates["quota"] = *account.Quota
	}
	if account.ExpireTime != nil {
		updates["expire_time"] = *account.ExpireTime
	}
	if account.Download != nil {
		updates["download"] = *account.Download
	}
	if account.Upload != nil {
		updates["upload"] = *account.Upload
	}
	if account.DeviceNo != nil {
		updates["device_no"] = *account.DeviceNo
	}
	if account.Deleted != nil {
		updates["deleted"] = *account.Deleted
	}
	if account.LoginAt != nil && *account.LoginAt > 0 {
		updates["login_at"] = *account.LoginAt
	}
	if account.ConAt != nil && *account.ConAt > 0 {
		updates["con_at"] = *account.ConAt
	}
	return dao.UpdateAccount([]int64{*account.Id}, updates)
}

func ResetTraffic(id int64) error {
	return dao.UpdateAccount([]int64{id}, map[string]interface{}{"download": 0, "upload": 0})
}

func ExistAccountUsername(username string, id int64) bool {
	var err error
	if id != 0 {
		_, err = dao.GetAccount("username = ? and id != ?", username, id)
	} else {
		_, err = dao.GetAccount("username = ?", username)
	}
	if err != nil {
		if err.Error() == constant.WrongPassword {
			return false
		}
	}
	return true
}

func GetAccount(id int64) (entity.Account, error) {
	return dao.GetAccount("id = ?", id)
}

func ListExportAccount() ([]bo.AccountExport, error) {
	accounts, err := dao.ListAccount(nil, nil)
	if err != nil {
		return nil, errors.New(constant.SysError)
	}
	var accountExports []bo.AccountExport
	for _, item := range accounts {
		accountExport := bo.AccountExport{
			Id:           *item.Id,
			Username:     *item.Username,
			Pass:         *item.Pass,
			ConPass:      *item.ConPass,
			Quota:        *item.Quota,
			Download:     *item.Download,
			Upload:       *item.Upload,
			ExpireTime:   *item.ExpireTime,
			DeviceNo:     *item.DeviceNo,
			KickUtilTime: *item.KickUtilTime,
			Role:         *item.Role,
			Deleted:      *item.Deleted,
			CreateTime:   *item.CreateTime,
			UpdateTime:   *item.UpdateTime,
			LoginAt:      *item.LoginAt,
			ConAt:        *item.ConAt,
		}
		accountExports = append(accountExports, accountExport)
	}
	return accountExports, nil
}

func ReleaseKickAccount(id int64) error {
	return dao.UpdateAccount([]int64{id}, map[string]interface{}{"kick_util_time": 0})
}

func UpsertAccount(accounts []entity.Account) error {
	return dao.UpsertAccount(accounts)
}

func GetAccountInfo(c *gin.Context) (vo.AccountInfoVo, error) {
	myClaims, err := ParseToken(GetToken(c))
	if err != nil {
		return vo.AccountInfoVo{}, err
	}
	if myClaims.AccountBo.Deleted != 0 {
		return vo.AccountInfoVo{}, errors.New("this account has been disabled")
	}
	return vo.AccountInfoVo{
		Id:       myClaims.AccountBo.Id,
		Username: myClaims.AccountBo.Username,
		Roles:    myClaims.AccountBo.Roles,
	}, nil
}
