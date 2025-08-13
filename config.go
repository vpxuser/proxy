package proxy

type Config struct {
	limiter      Limiter
	Negotiator   Negotiator
	Resolver     Resolver
	Dispatcher   Dispatcher
	TLSConfig    TLSConfig
	HttpHandler  HttpHandler
	WsHandler    WsHandler
	TcpHandler   TcpHandler
	dialer       Dialer
	DefaultSAN   string
	reqHandlers  []ReqHandlerFn
	respHandlers []RespHandlerFn
	wsHandlers   []WsHandlerFn
	rawHandlers  []RawHandlerFn
}

func NewConfig() *Config {
	return &Config{
		Negotiator:  defaultNegotiator,
		Resolver:    defaultResolver,
		Dispatcher:  defaultDispatcher,
		HttpHandler: defaultHttpHandler,
		WsHandler:   defaultWsHandler,
		TcpHandler:  defaultTcpHandler,
	}
}

type ConfigOption func(*Config)

func (c *Config) WithOptions(opts ...ConfigOption) *Config {
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func WithLimiter(limiter Limiter) ConfigOption {
	return func(c *Config) { c.limiter = limiter }
}

func WithNegotiator(negotiator Negotiator) ConfigOption {
	return func(c *Config) { c.Negotiator = negotiator }
}

func WithDispatcher(dispatcher Dispatcher) ConfigOption {
	return func(c *Config) { c.Dispatcher = dispatcher }
}

func WithTLSConfigFn(tlsConfig TLSConfig) ConfigOption {
	return func(c *Config) { c.TLSConfig = tlsConfig }
}

func WithHttpHandler(httpHandler HttpHandler) ConfigOption {
	return func(c *Config) { c.HttpHandler = httpHandler }
}

func WithWsHandler(wsHandler WsHandler) ConfigOption {
	return func(c *Config) { c.WsHandler = wsHandler }
}

func WithTcpHandler(tcpHandler TcpHandler) ConfigOption {
	return func(c *Config) { c.TcpHandler = tcpHandler }
}

func WithDialer(dialer Dialer) ConfigOption {
	return func(c *Config) { c.dialer = dialer }
}
