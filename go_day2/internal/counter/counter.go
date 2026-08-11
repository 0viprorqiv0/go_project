package counter

// EventCounter keeps the number of events seen for each IP address.
type EventCounter struct {
	counts map[string]int
}

// New creates an empty event counter.
func New() *EventCounter {
	return &EventCounter{counts: make(map[string]int)}
}

// Add increments and returns the count for ip.
// Empty IP addresses are ignored.
func (c *EventCounter) Add(ip string) int {
	if ip == "" {
		return 0
	}

	c.counts[ip]++
	return c.counts[ip]
}

// Count returns the current count for ip and whether it has been seen.
func (c *EventCounter) Count(ip string) (int, bool) {
	if ip == "" {
		return 0, false
	}

	count, exists := c.counts[ip]
	return count, exists
}
