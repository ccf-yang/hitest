package i18n

import (
	"embed"
	"fmt"
	"strings"
	"sync/atomic"

	"hitest/internal/core/app/repo"
	"hitest/internal/core/global"
	"hitest/internal/shared/ctl_conf"

	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

const defaultLang = "en"

var langFiles = map[string]string{
	"zh": "lang/zh.yaml",
}

func GetMsgWithMap(key string, maps map[string]interface{}) string {
	var content string
	if maps == nil {
		content, _ = global.I18n.Localize(&i18n.LocalizeConfig{
			MessageID: key,
		})
	} else {
		content, _ = global.I18n.Localize(&i18n.LocalizeConfig{
			MessageID:    key,
			TemplateData: maps,
		})
	}
	content = strings.ReplaceAll(content, ": <no value>", "")
	if content == "" {
		return key
	} else {
		return content
	}
}

func GetErrMsg(key string, maps map[string]interface{}) string {
	var content string
	if maps == nil {
		content, _ = global.I18n.Localize(&i18n.LocalizeConfig{
			MessageID: key,
		})
	} else {
		content, _ = global.I18n.Localize(&i18n.LocalizeConfig{
			MessageID:    key,
			TemplateData: maps,
		})
	}
	return content
}

func GetMsgByKey(key string) string {
	content, _ := global.I18n.Localize(&i18n.LocalizeConfig{
		MessageID: key,
	})
	return content
}

func Get(key string) string {
	content, _ := global.I18n.Localize(&i18n.LocalizeConfig{
		MessageID: key,
	})
	if content != "" {
		return content
	}
	return key
}

func GetWithName(key string, name string) string {
	var (
		dataMap = make(map[string]interface{})
	)
	dataMap["name"] = name
	content, _ := global.I18n.Localize(&i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: dataMap,
	})
	return content
}

func GetWithNameAndErr(key string, name string, err error) string {
	var (
		dataMap = make(map[string]interface{})
	)
	dataMap["name"] = name
	dataMap["err"] = err.Error()
	content, _ := global.I18n.Localize(&i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: dataMap,
	})
	return content
}

//go:embed lang/*
var fs embed.FS
var bundle *i18n.Bundle

func UseI18n() gin.HandlerFunc {
	return func(context *gin.Context) {
		lang := context.GetHeader("Accept-Language")
		if lang == "" {
			lang = GetLanguage()
		}
		global.I18n = i18n.NewLocalizer(bundle, lang)
	}
}

func Init() {
	if bundle == nil {
		initBundle()
	}
	dbLang := getLanguageFromDBInternal()
	if dbLang == "" {
		dbLang = defaultLang
	}
	SetCachedDBLanguage(dbLang)

	global.I18n = i18n.NewLocalizer(bundle, dbLang)
}

func UseI18nForCmd(lang string) {
	if bundle == nil {
		initBundle()
	}
	if lang == "" {
		langFromHitestctl := getLanguageFromHitestctl()
		if langFromHitestctl == "" {
			lang = defaultLang
		} else {
			lang = langFromHitestctl
		}
	}
	global.I18nForCmd = i18n.NewLocalizer(bundle, lang)
}

func GetMsgByKeyForCmd(key string) string {
	if global.I18nForCmd == nil {
		UseI18nForCmd("")
	}
	content, _ := global.I18nForCmd.Localize(&i18n.LocalizeConfig{
		MessageID: key,
	})
	return content
}

func GetMsgWithMapForCmd(key string, maps map[string]interface{}) string {
	if global.I18nForCmd == nil {
		UseI18nForCmd("")
	}
	var content string
	if maps == nil {
		content, _ = global.I18nForCmd.Localize(&i18n.LocalizeConfig{
			MessageID: key,
		})
	} else {
		content, _ = global.I18nForCmd.Localize(&i18n.LocalizeConfig{
			MessageID:    key,
			TemplateData: maps,
		})
	}
	content = strings.ReplaceAll(content, ": <no value>", "")
	if content == "" {
		return key
	} else {
		return content
	}
}

func getLanguageFromDBInternal() string {
	if global.DB == nil {
		return defaultLang
	}
	lang, _ := repo.NewISettingRepo().GetValueByKey("Language")
	if lang == "" {
		return defaultLang
	}
	return lang
}
func getLanguageFromHitestctl() string {
	if info := ctl_conf.LoadWithoutPanic("LANGUAGE"); info != "" && info != `""` {
		return info
	}
	return ""
}

var cachedDBLang atomic.Value

func GetLanguage() string {
	if v := cachedDBLang.Load(); v != nil {
		return v.(string)
	}
	return defaultLang
}

func SetCachedDBLanguage(lang string) {
	if lang == "" {
		lang = defaultLang
	}
	cachedDBLang.Store(lang)
}

func initBundle() {
	bundle = i18n.NewBundle(language.Chinese)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	isSuccess := true
	for _, file := range langFiles {
		if _, err := bundle.LoadMessageFileFS(fs, file); err != nil {
			fmt.Printf("[i18n] load language file %s failed: %v\n", file, err)
			isSuccess = false
		}
	}

	if !isSuccess {
		panic("[i18n] failed to init language files, See log above for details")
	}
}
