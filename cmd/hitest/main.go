// hitest: 一个二进制, 三种角色。
//
//	hitest serve                  中心服务（含内置本机 agent）
//	hitest agent -s ws://hub -t … 节点（NAS/Armbian 上运行）
//	hitest login / con / exec / cp / ls   CLI 客户端
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"hitest/internal/agent"
	"hitest/internal/client"
	"hitest/internal/server"
)

var version = "0.1.0"

func main() {
	root := &cobra.Command{
		Use:   "hitest",
		Short: "家庭/边缘 Linux 机器管理平台",
	}
	root.AddCommand(
		serveCmd(), agentCmd(),
		loginCmd(), conCmd(), execCmd(), cpCmd(), lsCmd(),
		versionCmd(),
	)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

// ---------- 角色 1: serve ----------

func serveCmd() *cobra.Command {
	var (
		configPath = "config.yaml"
		debug      = false
	)
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "启动中心服务",
		RunE: func(c *cobra.Command, args []string) error {
			cfg, err := server.LoadConfig(configPath)
			if err != nil {
				return err
			}
			if debug {
				cfg.Debug = true // 命令行优先; config.yaml 里写 debug: true 也一样
			}
			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()
			return server.Run(ctx, *cfg)
		},
	}
	cmd.Flags().StringVar(&configPath, "config", "config.yaml", "配置文件路径")
	cmd.Flags().BoolVar(&debug, "debug", false, "每 60s 打一行内存/Goroutine 快照, 并在 127.0.0.1:6061 开 pprof")
	return cmd
}

// ---------- 角色 2: agent ----------

func agentCmd() *cobra.Command {
	var (
		server = ""
		token  = ""
		name   = ""
		debug  = false
	)
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "以节点模式运行（装在 NAS/Armbian 上, 主动外连 server）",
		RunE: func(c *cobra.Command, args []string) error {
			if server == "" || token == "" {
				return fmt.Errorf("必须提供 -s <server-url> 和 -t <token>")
			}
			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()
			return agent.Start(ctx, agent.Options{ID: name, Token: token, Server: server, Debug: debug})
		},
	}
	cmd.Flags().StringVarP(&server, "server", "s", "", "server 地址 ws://host:port")
	cmd.Flags().StringVarP(&token, "token", "t", "", "注册 token")
	cmd.Flags().StringVarP(&name, "name", "n", "", "机器 ID (默认 hostname)")
	cmd.Flags().BoolVar(&debug, "debug", false, "每 60s 打一行内存/Goroutine 快照, 并在 127.0.0.1:6062 开 pprof")
	return cmd
}

// ---------- 角色 3: CLI 客户端 ----------

func loginCmd() *cobra.Command {
	var token string
	cmd := &cobra.Command{
		Use:   "login <server-url>",
		Short: "保存 server 地址与 token",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			cfg := &client.Config{Server: args[0], Token: token}
			if err := client.Save(cfg); err != nil {
				return err
			}
			path, _ := client.ConfigPath()
			fmt.Println("✅ 已保存", path)
			return nil
		},
	}
	cmd.Flags().StringVarP(&token, "token", "t", "", "access token")
	return cmd
}

func conCmd() *cobra.Command {
	var cmdName string
	cmd := &cobra.Command{
		Use:   "con <machine>",
		Short: "打开交互终端（像 ssh）",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			cfg, err := client.Load()
			if err != nil {
				return err
			}
			return cfg.Con(args[0], cmdName)
		},
	}
	cmd.Flags().StringVar(&cmdName, "cmd", "/bin/bash -l", "远端启动的 shell")
	return cmd
}

func execCmd() *cobra.Command {
	var (
		timeout = 30
		asJSON  = false
	)
	cmd := &cobra.Command{
		Use:   "exec <machine> <command>",
		Short: "在机器上执行命令（非交互, 适合脚本与 AI agent）",
		Args:  cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			cfg, err := client.Load()
			if err != nil {
				return err
			}
			return cfg.Exec(args[0], args[1], timeout, asJSON)
		},
	}
	cmd.Flags().IntVar(&timeout, "timeout", 30, "超时秒数")
	cmd.Flags().BoolVar(&asJSON, "json", false, "结构化输出")
	return cmd
}

func cpCmd() *cobra.Command {
	var recursive bool
	cmd := &cobra.Command{
		Use:   "cp <src> <machine1[,machine2...]:/dst/path>",
		Short: "上传文件/目录到一台或多台机器",
		Args:  cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			cfg, err := client.Load()
			if err != nil {
				return err
			}
			return cfg.Cp(args[0], args[1], recursive)
		},
	}
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "递归上传目录")
	return cmd
}

func lsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls <machine> [path]",
		Short: "列出机器上的目录",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(c *cobra.Command, args []string) error {
			cfg, err := client.Load()
			if err != nil {
				return err
			}
			path := "/"
			if len(args) == 2 {
				path = args[1]
			}
			return cfg.Ls(args[0], path)
		},
	}
	return cmd
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "显示版本",
		Run: func(c *cobra.Command, args []string) {
			fmt.Println("hitest", version)
		},
	}
}
