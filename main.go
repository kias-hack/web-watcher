package main

import (
    "flag"
    "os"
    "github.com/BurntSushi/toml"
    "sync"
)

type Telegram struct {
    BotToken string `toml:"bot_token"`
    ChatID int64 `toml:"chat_id"`
}

type Email struct {
    Email string `toml:"email"`
}

type Notification struct {
    Type string `toml:"type"`

    Telegram
    Email
}

type ScrapeConfig struct {
    ReadTimeOut int `toml:"read_timeout"`
    FollowRedirectsCount bool `toml:"follow_redirects_count"`
    ScrapePeriod: time.Duration `toml:'scrape_period'`
}

type Service struct {
    ScrapeConfig
    Name string `toml:"name"`
    URL string `toml:"url"`
}

type Config struct {
    ScrapeConfig ScrapeConfig `toml:"scrape_config"`
    Services []Service `toml:"services"`
    Notification Notification `toml:"notification"`
}


type ObserverManager struct {
    Config Config
    
    ctx context.Context
    cancel context.CancelFunc
    wg *sync.WaitGroup
}

func NewObserverManager(ctx context.Context, config Config) *ObserverManager {
    ctx, cancel := context.WithCancel(ctx)
    
    return &ObserverManager{
        ctx: ctx,
        cancel: cancel,
        wg: &sync.WaitGroup{},
        Config: config,
    }
}

func (o *ObserverManager) Start() {
    for _, service := range o.Config.Services {
        wg.Add(1)
        go o.scrapeService(service)
    }
}

func (o *ObserverManager) scrapeService(wg *sync.WaitGroup, service Service) {
    defer wg.Done()

    ticker, cancel := time.NewTicker(service.ScrapePeriod)
    defer cancel()

    defer 

    for {
        select {
            case <-o.ctx.Done():
                return
            case <-ticker.C:
                // TODO scrape service
        }
    }
}

func (o *ObserverManager) Stop(ctx context.Context) error {
    o.cancel()

    waitCancelCh := make(chan strct{}, 1)

    go func(){
        o.wg.Wait()

        close(waitCancelCh)
    }()

    select {
        case <-ctx.Done():
            close(waitCancelCh)
            return fmt.Errorf("context done, while stopping server handlers: %w", ctx.Err())
        case <-waitCancelCh:
    }

    return nil
}

func main() {
    var configPath string
    flag.StringVar(&configPath, "config", "config.yaml", "path to config file")

    flag.Parse()

    slog.Info("starting application initialization with config path: %s", configPath)

    config := mustGetConfig(configPath)
}

func mustGetConfig(configPath string) Config {
    if configPath == "" {
        panic("config path is required")
    }

    data, err :=os.ReadFile(configPath)
    if err != nil {
        panic(fmt.Errorf("failed to read config file: %w", err))
    }

    var config Config

    if err := toml.Decode(data, &config); err != nil {
        panic(fmt.Errorf("failed to decode config file: %w", err))
    }

    // TODO проверить поля Service 

    return config
}