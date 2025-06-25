package proxy

type Config struct {
	limiter      Limiter
	negotiator   Negotiator
	dispatcher   Dispatcher
	tlsConfig    TLSConfig
	httpHandler  HttpHandler
	wsHandler    WsHandler
	tcpHandler   TcpHandler
	dialer       Dialer
	DefaultSAN   string
	reqHandlers  []ReqHandlerFn
	respHandlers []RespHandlerFn
	wsHandlers   []WsHandlerFn
	rawHandlers  []RawHandlerFn
}

func NewConfig() *Config {
	return &Config{
		dispatcher:  defaultDispatcher,
		httpHandler: defaultHttpHandler,
		wsHandler:   defaultWsHandler,
		tcpHandler:  defaultTcpHandler,
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
	return func(c *Config) { c.negotiator = negotiator }
}

func WithDispatcher(dispatcher Dispatcher) ConfigOption {
	return func(c *Config) { c.dispatcher = dispatcher }
}

func WithTLSConfigFn(tlsConfig TLSConfig) ConfigOption {
	return func(c *Config) { c.tlsConfig = tlsConfig }
}

func WithHttpHandler(httpHandler HttpHandler) ConfigOption {
	return func(c *Config) { c.httpHandler = httpHandler }
}

func WithWsHandler(wsHandler WsHandler) ConfigOption {
	return func(c *Config) { c.wsHandler = wsHandler }
}

func WithTcpHandler(tcpHandler TcpHandler) ConfigOption {
	return func(c *Config) { c.tcpHandler = tcpHandler }
}

func WithDialer(dialer Dialer) ConfigOption {
	return func(c *Config) { c.dialer = dialer }
}
