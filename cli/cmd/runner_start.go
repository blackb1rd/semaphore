package cmd

import (
	"github.com/semaphoreui/semaphore/util"
	"github.com/spf13/cobra"
)

var runnerStartArgs struct {
	register bool
}

func init() {
	runnerStartCmd.PersistentFlags().BoolVar(&runnerStartArgs.register, "register", false, "Register new runner if not registered")
	runnerCmd.AddCommand(runnerStartCmd)
}

func runRunner() {

	configFile := util.ConfigInit(persistentFlags.configPath, persistentFlags.noConfig)

	taskPool := createRunnerJobPool()

	if runnerStartArgs.register {

		initRunnerRegistrationToken()

		if util.Config.Runner.Token == "" {

			err := taskPool.Register(configFile)

			if err != nil {
				panic(err)
			}
		}
	}

	taskPool.Run()
}

var runnerStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Run in runner mode",
	Run: func(cmd *cobra.Command, args []string) {
		runRunner()
	},
}
