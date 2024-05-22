/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/hirosassa/zerodriver"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	telebot "gopkg.in/telebot.v3"
)

var (
	TeleToken   = os.Getenv("TELE_TOKEN")
	MetricsHost = os.Getenv("METRICS_HOST")
)

func initMetrics(ctx context.Context) {

	exporter, _ := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithEndpoint(MetricsHost),
		otlpmetricgrpc.WithInsecure(),
	)

	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(fmt.Sprintf("hobot_%s", appVersion)),
	)

	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(resource),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(10*time.Second)),
		),
	)

	otel.SetMeterProvider(mp)
}

func pmetrics(ctx context.Context, payload string) {

	meter := otel.GetMeterProvider().Meter("hobot_ligth_signal_counter")

	counter, _ := meter.Int64Counter(fmt.Sprintf("hobot_ligth_signal_%s", payload))

	counter.Add(ctx, 1)
}

// hobotCmd represents the hobot command
var hobotCmd = &cobra.Command{
	Use:     "hobot",
	Aliases: []string{"start"},
	Short:   "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := zerodriver.NewProductionLogger()

		hobot, err := telebot.NewBot(telebot.Settings{
			URL:    "",
			Token:  TeleToken,
			Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
		})

		if err != nil {
			logger.Fatal().Str("Error", err.Error()).Msg("Please check TELE_TOKEN")
			return
		} else {
			logger.Info().Str("Version", appVersion).Msg("hobot running")

		}

		trafficSignal := make(map[string]map[string]int8)

		trafficSignal["red"] = make(map[string]int8)
		trafficSignal["amber"] = make(map[string]int8)
		trafficSignal["green"] = make(map[string]int8)

		trafficSignal["red"]["pin"] = 12
		trafficSignal["amber"]["pin"] = 27
		trafficSignal["green"]["pin"] = 22

		hobot.Handle(telebot.OnText, func(m telebot.Context) error {
			logger.Info().Str("Payload", m.Text()).Msg(m.Message().Payload)

			payload := m.Message().Payload
			pmetrics(context.Background(), payload)

			switch payload {
			case "hello":
				err = m.Send(fmt.Sprintf("Hello, my name is HoBot %s!", appVersion))

			case "red", "amber", "green":

				if trafficSignal[payload]["on"] == 0 {
					trafficSignal[payload]["on"] = 1
				} else {
					trafficSignal[payload]["on"] = 0
				}

				err = m.Send(fmt.Sprintf("Switch %s ligth signal to %d", payload, trafficSignal[payload]["on"]))

			default:
				err = m.Send("Usage: /s red|amber|green")

			}
			return err

		})

		hobot.Start()
	},
}

func init() {
	ctx := context.Background()
	initMetrics(ctx)
	rootCmd.AddCommand(hobotCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// hobotCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// hobotCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
