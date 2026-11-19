// Package module provides internal types and functions for the GoBE application.
package module

import (
	"github.com/kubex-ecosystem/ethyr/internal/module/style"
	"github.com/kubex-ecosystem/ethyr/internal/module/version"
	"github.com/kubex-ecosystem/ethyr/tools/info"
	"github.com/spf13/cobra"
)

// Ethyr is a struct that implements the Module interface.
type Ethyr struct {
	parentCmdName string
	hideBanner    bool
	certPath      string
	keyPath       string
	configPath    string
	Banners       []string
}

// Alias returns the alias for the module.
func (m *Ethyr) Alias() string {
	return ""
}

// ShortDescription returns the short description for the module.
func (m *Ethyr) ShortDescription() string {
	return "Domus: GKBX Database and Docker manager/service. "
}

// LongDescription returns the long description for the module.
func (m *Ethyr) LongDescription() string {
	return `Domus: Is a tool to manage GKBX database and Docker services. It provides many DB flavors like MySQL, PostgreSQL, MongoDB, Redis, etc. It also provides Docker services like Docker Swarm, Docker Compose, etc. It is a command line tool that can be used to manage GKBX database and Docker services.`
}

// Usage returns the usage for the module.
func (m *Ethyr) Usage() string {
	return "domus [command] [args]"
}

// Examples returns the examples for the module.
func (m *Ethyr) Examples() []string {
	return []string{"domus [command] [args]", "domus database user auth'", "domus db roles list"}
}

// Active returns true if the module is active.
func (m *Ethyr) Active() bool {
	return true
}

// Module returns the module name.
func (m *Ethyr) Module() string {
	return "domus"
}

// Execute executes the module.
func (m *Ethyr) Execute() error {
	dbChanData := make(chan interface{})
	defer close(dbChanData)

	if spyderErr := m.Command().Execute(); spyderErr != nil {
		return spyderErr
	}
	return nil
}

// Command returns the cobra command for the module.
func (m *Ethyr) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use: m.Module(),
		//Aliases:     []string{m.Alias(), "w", "wb", "webServer", "http"},
		Example: m.concatenateExamples(),
		Annotations: info.CLIBannerStyle(
			m.Banners,
			[]string{
				m.LongDescription(),
				m.ShortDescription(),
			},
			m.hideBanner,
		),
		Version: version.GetVersion(),
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	cmd.AddCommand(version.CliCommand())

	style.SetUsageTemplate(cmd)

	return cmd
}

// SetParentCmdName sets the parent command name.
func (m *Ethyr) SetParentCmdName(rtCmd string) {
	m.parentCmdName = rtCmd
}

// concatenateExamples concatenates the examples for the module.
func (m *Ethyr) concatenateExamples() string {
	examples := ""
	rtCmd := m.parentCmdName
	if rtCmd != "" {
		rtCmd = rtCmd + " "
	}
	for _, example := range m.Examples() {
		examples += rtCmd + example + "\n  "
	}
	return examples
}
