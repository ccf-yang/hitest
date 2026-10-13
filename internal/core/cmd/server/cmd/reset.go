package cmd

import (
	"fmt"

	"hitest/internal/core/i18n"
	"hitest/internal/shared/constant"

	"github.com/spf13/cobra"
)

func init() {
	resetCmd.SetHelpFunc(func(c *cobra.Command, s []string) {
		i18n.UseI18nForCmd(language)
		loadResetHelper()
	})

	RootCmd.AddCommand(resetCmd)
	resetCmd.AddCommand(resetSSLCmd)
	resetCmd.AddCommand(resetEntranceCmd)
	resetCmd.AddCommand(resetBindIpsCmd)
	resetCmd.AddCommand(resetDomainCmd)
}

var resetCmd = &cobra.Command{
	Use: "reset",
	RunE: func(cmd *cobra.Command, args []string) error {
		i18n.UseI18nForCmd(language)
		loadResetHelper()
		return nil
	},
}

var resetSSLCmd = &cobra.Command{
	Use: "https",
	RunE: func(cmd *cobra.Command, args []string) error {
		i18n.UseI18nForCmd(language)
		if !isRoot() {
			fmt.Println(i18n.GetMsgWithMapForCmd("SudoHelper", map[string]interface{}{"cmd": "sudo hitestctl reset https"}))
			return nil
		}
		db, err := loadDBConn("core.db")
		if err != nil {
			return err
		}

		if err := setSettingByKey(db, "SSL", constant.StatusDisable); err != nil {
			return err
		}
		return nil
	},
}
var resetEntranceCmd = &cobra.Command{
	Use: "entrance",
	RunE: func(cmd *cobra.Command, args []string) error {
		i18n.UseI18nForCmd(language)
		if !isRoot() {
			fmt.Println(i18n.GetMsgWithMapForCmd("SudoHelper", map[string]interface{}{"cmd": "sudo hitestctl reset entrance"}))
			return nil
		}
		db, err := loadDBConn("core.db")
		if err != nil {
			return err
		}

		return setSettingByKey(db, "SecurityEntrance", "")
	},
}
var resetBindIpsCmd = &cobra.Command{
	Use: "ips",
	RunE: func(cmd *cobra.Command, args []string) error {
		i18n.UseI18nForCmd(language)
		if !isRoot() {
			fmt.Println(i18n.GetMsgWithMapForCmd("SudoHelper", map[string]interface{}{"cmd": "sudo hitestctl reset ips"}))
			return nil
		}
		db, err := loadDBConn("core.db")
		if err != nil {
			return err
		}

		if err := setSettingByKey(db, "AllowIPs", ""); err != nil {
			return err
		}
		return nil
	},
}
var resetDomainCmd = &cobra.Command{
	Use: "domain",
	RunE: func(cmd *cobra.Command, args []string) error {
		i18n.UseI18nForCmd(language)
		if !isRoot() {
			fmt.Println(i18n.GetMsgWithMapForCmd("SudoHelper", map[string]interface{}{"cmd": "sudo hitestctl reset domain"}))
			return nil
		}
		db, err := loadDBConn("core.db")
		if err != nil {
			return err
		}

		if err := setSettingByKey(db, "BindDomain", ""); err != nil {
			return err
		}
		return nil
	},
}

func loadResetHelper() {
	fmt.Println(i18n.GetMsgByKeyForCmd("ResetCommands"))
	fmt.Println("\nUsage:\n  hitest reset [command]\n\nAvailable Commands:")
	fmt.Println("\n  domain      " + i18n.GetMsgByKeyForCmd("ResetDomain"))
	fmt.Println("  entrance    " + i18n.GetMsgByKeyForCmd("ResetEntrance"))
	fmt.Println("  https       " + i18n.GetMsgByKeyForCmd("ResetHttps"))
	fmt.Println("  ips         " + i18n.GetMsgByKeyForCmd("ResetIPs"))
	fmt.Println("\nFlags:\n  -h, --help   help for reset")
	fmt.Println("\nUse \"hitest reset [command] --help\" for more information about a command.")
}
