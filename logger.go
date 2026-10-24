package tradingdata

var Logger func(msg string) = func(msg string) {}

func (c *Client) SetLogger(f func(msg string)) {
	if f != nil {
		Logger = f
	}
}
