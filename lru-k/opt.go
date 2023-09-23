package lru_k

type Option func(*cache)

func WithK(k int) Option {
	return func(c *cache) {
		c.k = k
	}
}

func WithMaxItem(maxItem int) Option {
	return func(c *cache) {
		c.maxItem = maxItem
	}
}

func WithInactiveLimit(inactiveLimit int) Option {
	return func(c *cache) {
		c.inactiveLimit = inactiveLimit
	}
}

func WithOnEliminate(onEliminate func(k string, v any)) Option {
	return func(c *cache) {
		c.onEliminate = onEliminate
	}
}
