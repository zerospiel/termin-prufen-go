package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"strconv"
	"time"

	"github.com/zerospiel/termin-prufen-go/pkg/prufen"
	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v3"
)

func main() {
	l := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				source := a.Value.Any().(*slog.Source)
				// dir/file:#
				a.Value = slog.AnyValue(fmt.Sprintf("%s:%d",
					filepath.Join(
						filepath.Base(filepath.Dir(source.File)),
						filepath.Base(source.File)),
					source.Line))
				a.Key = "caller"
			}

			return a
		},
	}))
	slog.SetDefault(l)

	cfg, err := getConfig()
	if err != nil {
		l.Error("failed to evaluate config", "error", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	options := prufen.Options{
		TelegramAPIToken: cfg.TelegramBotToken,
		TelegramChatID:   cfg.TelegramChatID,

		Citizenship:             cfg.Citizenship,
		PeopleNumber:            cfg.PeopleNumber,
		LiveInBerlin:            cfg.LiveInBerlin,
		FamilyMemberCitizenship: cfg.FamilyMemberCitizenship,
		Reason:                  cfg.Reason,

		ScenarioTimeout:         cfg.ScenarioTimeout,
		ScreenshotsPath:         cfg.ScreenshotsDir,
		PollInterval:            cfg.PollInterval,
		GracefulShutdownTimeout: cfg.GracefulShutdownTimeout,
		Port:                    cfg.Port,
	}
	if cfg.Debug {
		options.DebugFunc = l.Debug
	}

	runner, err := prufen.NewRunner(options)
	if err != nil {
		l.Error("failed to init runner", "error", err)
		return
	}

	if cfg.SingleRunMode {
		l.Info("Running in a single mode")
		runner.RunFullCycle()
		return
	}

	if err := runner.Run(ctx); err != nil {
		l.Error("failed to run, shutting down...", "error", err)
		return
	}
}

type (
	Config struct {
		*AbhConfig      `yaml:",inline,omitempty"`
		*TelegramConfig `yaml:",inline,omitempty"`
		*AppConfig      `yaml:",inline,omitempty"`
	}

	AbhConfig struct {
		Citizenship             string `yaml:"citizenship,omitempty"`
		PeopleNumber            string `yaml:"people_number,omitempty"`
		LiveInBerlin            string `yaml:"live_in_berlin,omitempty"`
		FamilyMemberCitizenship string `yaml:"family_member_citizenship,omitempty"`
		Reason                  string `yaml:"reason,omitempty"`
	}

	TelegramConfig struct {
		TelegramBotToken string `yaml:"telegram_bot_token,omitempty"`
		TelegramChatID   int64  `yaml:"telegram_chat_id,omitempty"`
	}

	AppConfig struct {
		ConfigFile              string
		ScreenshotsDir          string        `yaml:"screenshots_dir,omitempty"`
		Port                    int           `yaml:"port,omitempty"`
		ScenarioTimeout         time.Duration `yaml:"scenario_timeout,omitempty"`
		PollInterval            time.Duration `yaml:"poll_interval,omitempty"`
		GracefulShutdownTimeout time.Duration `yaml:"graceful_shutdown_timeout,omitempty"`

		SingleRunMode bool `yaml:"single_run_mode,omitempty"`
		Debug         bool `yaml:"debug,omitempty"`
	}
)

func getConfig() (*Config, error) {
	debug := flag.Bool("debug", false, "Print debug logs from Chrome to the stdout stream")
	singleMode := flag.Bool("single-run-mode", false, "Run the application only once. Could be useful for test purposes or to develop more automations")

	var configFile string
	flag.StringVar(&configFile, "config-file", "", "Config file with settings")
	flag.Parse()

	if configFile == "" {
		var (
			configFileEnv = "CONFIG_FILE"
			ok            bool
		)
		configFile, ok = os.LookupEnv(configFileEnv)
		if !ok {
			return nil, fmt.Errorf("either %q flag and %q env not presented, please provide it", "--config-file", configFileEnv)
		}
	}

	configFileAbs, err := filepath.Abs(configFile)
	if err != nil {
		return nil, fmt.Errorf("unable to get abs path for %q: %v", configFile, err)
	}

	bb, err := os.ReadFile(configFileAbs)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %v", configFileAbs, err)
	}

	cfg := Config{
		AppConfig: &AppConfig{
			SingleRunMode: *singleMode,
			Debug:         *debug,
		},
	}

	if err := yaml.Unmarshal(bb, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal file %q: %v", configFileAbs, err)
	}

	if reflect.ValueOf(cfg.TelegramConfig).IsZero() {
		return nil, fmt.Errorf("no telegram API credentials were given")
	}
	if cfg.TelegramChatID < 1 {
		return nil, fmt.Errorf("wrong value in param \"telegram_chat_id\" %d, should be more than 1", cfg.TelegramChatID)
	}

	screenshotDirAbs, err := filepath.Abs(cfg.ScreenshotsDir)
	if err != nil {
		return nil, fmt.Errorf("failedto get abs path for %q: %v", cfg.ScreenshotsDir, err)
	}
	cfg.ScreenshotsDir = screenshotDirAbs

	// validate a little abh config
	if reflect.ValueOf(cfg.AbhConfig).IsZero() {
		return nil, fmt.Errorf("no ABH config were given")
	}
	if cfg.LiveInBerlin != "yes" &&
		cfg.LiveInBerlin != "no" {
		return nil, fmt.Errorf("param \"live_in_berlin\" can only be \"yes\" or \"no\", got %q", cfg.LiveInBerlin)
	}
	numApp, err := strconv.ParseInt(cfg.PeopleNumber, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse param \"people_number\" %q: %v", cfg.PeopleNumber, err)
	}
	if numApp < 1 || numApp > 8 {
		return nil, fmt.Errorf("wrong number in param \"people_number\" given %d, valid [1-8]", numApp)
	}
	// TODO: validate reason?

	return &cfg, nil
}
