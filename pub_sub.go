package cb_cache

type queue struct {
	keys chan string // expect to delete
}

// Publish is concurrent safety and may be blocked
func (q *queue) Publish(k string) {
	q.keys <- k
}

// subscribe is concurrent safety and may be blocked
func (q *queue) subscribe() string {
	return <-q.keys
}
