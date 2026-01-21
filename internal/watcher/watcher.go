package watcher

type Watcher struct {
    Config Config
    
    ctx context.Context
    cancel context.CancelFunc
    wg *sync.WaitGroup
}

func NewWatcher(ctx context.Context, config Config) *Watcher {
    ctx, cancel := context.WithCancel(ctx)
    
    return &Watcher{
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

    // for {
    //     select {
    //         case <-o.ctx.Done():
    //             return
    //         case <-ticker.C:
    //             // TODO scrape service
    //     }
    // }
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