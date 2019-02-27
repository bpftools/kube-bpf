package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/viper"
	"net/http"
	"os"
	"time"

	mapc "github.com/bpftools/kube-bpf/mapcollector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	logg *zap.Logger
	addr string
	coll *mapc.MapCollector
	tout = 2 * time.Second
)

var root = &cobra.Command{
	Use:   "runbpf <object file>",
	Short: "",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Setup logging
		logConf := []byte(`{
			"level": "debug",
			"encoding": "console",
			"outputPaths": ["stdout"],
			"errorOutputPaths": ["stderr"],
			"initialFields": {},
			"encoderConfig": {
			  "messageKey": "message",
			  "levelKey": "level",
			  "levelEncoder": "lowercase"
			}
		}`)
		var cfg zap.Config
		if err := json.Unmarshal(logConf, &cfg); err != nil {
			return err
		}
		var err error
		logg, err = cfg.Build()
		if err != nil {
			return err
		}

		// Setup a collector for input BPF program
		coll = mapc.New(args[0], logg)
		if err := coll.Setup(); err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		serve(ctx)

		if err := logg.Sync(); err != nil {
			return fmt.Errorf("error syncing logs: %v", err)
		}

		<-ctx.Done()
		return nil
	},
}

func main() {
	root.PersistentFlags().StringVarP(&addr, "address", "a", ":9387", "address of the server")
	viper.BindPFlags(root.PersistentFlags())

	var exitCode int
	if err := root.Execute(); err != nil {
		exitCode = 1
	}

	time.Sleep(tout)
	os.Exit(exitCode)
}

type registry struct {
	*prometheus.Registry

	logger *zap.Logger
}

type logger struct {
	r *registry
}

var _ promhttp.Logger = (*logger)(nil)

func (l logger) Println(v ...interface{}) {
	l.r.logger.Sugar().Info(v...)
}

func serve(ctx context.Context) {
	// By using DefaultServeMux profiling endpoints come for free.
	server := &http.Server{Addr: addr, Handler: http.DefaultServeMux}
	l := logg.With(zap.String("addr", addr))

	reg := &registry{
		Registry: prometheus.NewRegistry(),
		logger:   l.With(zap.String("service", "registry")),
	}
	// reg.MustRegister(prometheus.NewGoCollector())
	reg.MustRegister(coll)

	opt := promhttp.HandlerOpts{
		ErrorLog: logger{r: reg},
	}

	go func() {
		http.Handle("/metrics", promhttp.HandlerFor(reg, opt))
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			l.Error("Server error", zap.Error(err))
		}
		l.Info("Server closed")
	}()
	l.Info("Server started")
	<-ctx.Done()
	// Gracefully shutdown server.
	ctx, cancel := context.WithTimeout(context.Background(), tout)
	defer cancel()
	server.Shutdown(ctx)
}
