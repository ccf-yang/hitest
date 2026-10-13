package encrypt

import (
	"hitest/internal/agent/global"
	crypt "hitest/internal/shared/encrypt"
	"hitest/internal/shared/model"
)

func StringDecrypt(text string) (string, error) {
	defer func() {
		if r := recover(); r != nil {
			global.LOG.Errorf("A panic occurred during string decrypt with key, error message: %v", r)
		}
	}()
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
	return crypt.StringDecryptWithKey(text, global.CONF.Base.EncryptKey)
}
