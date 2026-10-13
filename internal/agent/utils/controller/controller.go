package controller

import (
	"errors"
	"os/exec"

	"hitest/internal/agent/global"
	"hitest/internal/agent/utils/controller/manager"
)

type Controller interface {
	Name() string
	IsActive(serviceName string) (bool, error)
	IsEnable(serviceName string) (bool, error)
	IsExist(serviceName string) (bool, error)
	Status(serviceName string) (string, error)

	Operate(operate, serviceName string) error

	Reload() error
}

func New() (Controller, error) {
	managerOptions := []string{"systemctl", "rc-service", "service"}
	for _, item := range managerOptions {
		if _, err := exec.LookPath(item); err != nil {
			continue
		}
		switch item {
		case "systemctl":
			return manager.NewSystemd(), nil
		case "rc-service":
			return manager.NewOpenrc(), nil
		case "service":
			return manager.NewSysvinit(), nil
		}
	}
	return nil, errors.New("not support such manager initializatio")
}

func RestartPanel(core, agent, reload bool) {
	client, err := New()
	if err != nil {
		global.LOG.Errorf("load client for controller failed, err: %v", err)
		return
	}
	if reload {
		if err := client.Reload(); err != nil {
			global.LOG.Errorf("restart hitest service failed, err: %v", err)
			return
		}
	}
	if core {
		if err := client.Operate("restart", "hitest-core"); err != nil {
			global.LOG.Errorf("restart hitest core service failed, err: %v", err)
			return
		}
	}
	if agent {
		if err := client.Operate("restart", "hitest-agent"); err != nil {
			global.LOG.Errorf("restart hitest agent service failed, err: %v", err)
			return
		}
	}
}

// GetServicePath returns the configuration file path for the specified service.
// If serviceName is empty, it returns the default directory path based on the system's service manager.
// For non-empty serviceName, it retrieves the exact service file path.
// Parameters:
//   - serviceName: Name of the service. If empty, returns the default directory.
//
// Returns:
//   - string: The service configuration file path.
//   - error: Error if the service manager is unsupported or command execution fails.
