package encrypt

import (
	"hitest/internal/core/global"
	crypt "hitest/internal/shared/encrypt"
	"hitest/internal/shared/model"
)

func StringEncrypt(text string) (string, error) {
	if len(text) == 0 {
		return "", nil
	}
	if len(global.CONF.Base.EncryptKey) == 0 {
		var encryptSetting model.Setting
		if err := global.DB.Where("key = ?", "EncryptKey").First(&encryptSetting).Error; err != nil {
			return "", err
		}
		global.CONF.Base.EncryptKey = encryptSetting.Value
	}
	return crypt.StringEncryptWithKey(text, global.CONF.Base.EncryptKey)
}
