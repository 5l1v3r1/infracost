package main

import (
	"fmt"
	"os"
	"time"

	"github.com/infracost/infracost/pkg/config"
	"github.com/infracost/infracost/pkg/output"
	"github.com/infracost/infracost/pkg/prices"
	"github.com/infracost/infracost/pkg/schema"

	"github.com/infracost/infracost/internal/providers/terraform"

	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	formatter := &log.TextFormatter{
		DisableTimestamp:       true,
		DisableLevelTruncation: true,
	}
	log.SetFormatter(formatter)

	app := &cli.App{
		Name:                 "infracost",
		Usage:                "Generate cost reports from Terraform plans",
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:      "tfjson",
				Usage:     "Path to Terraform Plan JSON file",
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:      "tfplan",
				Usage:     "Path to Terraform Plan file. Requires tfdir to also be set",
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:      "tfdir",
				Usage:     "Path to the Terraform project directory",
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:  "log-level",
				Usage: "Log level (TRACE, DEBUG, INFO, WARN, ERROR)",
				Value: "WARN",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output (json, table)",
				Value:   "table",
			},
			&cli.BoolFlag{
				Name:  "no-color",
				Usage: "Turn off colored output",
				Value: false,
			},
			&cli.StringFlag{
				Name:  "api-url",
				Usage: "Price List API URL",
			},
		},
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			log.Error(err)
			_ = cli.ShowAppHelp(c)
			os.Exit(1)
			return nil
		},
		Action: func(c *cli.Context) error {
			if c.Bool("no-color") {
				config.Config.NoColor = true
				formatter.DisableColors = true
				color.NoColor = true
			}

			if c.String("log-level") != "" {
				switch c.String("log-level") {
				case "TRACE":
					log.SetLevel(log.TraceLevel)
				case "DEBUG":
					log.SetLevel(log.DebugLevel)
				case "INFO":
					log.SetLevel(log.InfoLevel)
				case "WARN":
					log.SetLevel(log.WarnLevel)
				case "ERROR":
					log.SetLevel(log.ErrorLevel)
				}
			}

			if c.String("api-url") != "" {
				config.Config.ApiUrl = c.String("api-url")
			}

			provider := terraform.Provider()
			err := provider.ProcessArgs(c)
			if err != nil {
				color.HiRed(err.Error())
				_ = cli.ShowAppHelp(c)
				os.Exit(1)
			}

			s := spinner.New(spinner.CharSets[14], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
			if !c.Bool("no-color") {
				_ = s.Color("fgHiGreen", "bold")
			}
			s.Suffix = " Calculating costs…"
			s.Start()

			resources, err := provider.LoadResources()
			if err != nil {
				return err
			}
			err = prices.PopulatePrices(resources)
			if err != nil {
				return err
			}
			schema.CalculateCosts(resources)
			schema.SortResources(resources)

			var out []byte
			switch c.String("output") {
			case "table":
				out, err = output.ToTable(resources)
			case "json":
				out, err = output.ToJSON(resources)
			default:
				err = cli.ShowAppHelp(c)
			}
			if err != nil {
				return err
			}
			s.Stop()

			fmt.Println(string(out))

			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
