package app_with_helm_flag

import "github.com/urfave/cli"

type AppWithFlag struct {
	*cli.App
	kubeApiServer string
	kubeToken string
	kubeCAFile string
	kubeConfig string
}

func New(kubeConfig,kubeCAFile,kubeToken,kubeApiServer  string) *AppWithFlag  {
	return &AppWithFlag{
		App:cli.NewApp(),
		kubeApiServer:kubeApiServer,
		kubeToken:kubeToken,
		kubeCAFile:kubeCAFile,
		kubeConfig:kubeConfig,
	}
}
