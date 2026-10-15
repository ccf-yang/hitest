package commands

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hitest/internal/store"
)

type CommandsModule struct{ svc *QuickCommandService }

func (m *CommandsModule) Name() string { return "commands" }

func (m *CommandsModule) RegisterRoutes(api gin.IRouter, _ *store.Store) {
	m.svc = &QuickCommandService{}
	api.GET("/commands/preset", m.list)
	api.POST("/commands/preset/run", m.run)
}

func (m *CommandsModule) list(c *gin.Context) {
	commands, err := m.svc.ListQuickCommands()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, commands)
}

func (m *CommandsModule) run(c *gin.Context) {
	var req RunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := m.svc.RunQuickCommand(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"output": result})
}
