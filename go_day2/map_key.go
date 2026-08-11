package main

import "fmt"

type eventCounter struct {
	counts map[string]int
}

func newEventCounter() *eventCounter{
	return &eventCounter{
		counts: make(map[string]int),
	}
}

func(c *eventCounter) add(ip string) (int, bool){
	if c ==  nil || ip == "" {
		return 0, false
	}
	if c.counts == nil {
		c.counts = make(map[string]int)
	}
	c.counts[ip]++
	return c.counts[ip], true
}

func (c *eventCounter) count(ip string) (int, bool) {
	if c == nil {
		return 0, false
	}

	count, exists := c.counts[ip]
	return count, exists
}

func(c *eventCounter) remove(ip string) {
	if c == nil {
		return
	}
	delete(c.counts, ip)
}

func main(){
	counter := newEventCounter()
	first, added := counter.add("192.0.2.10")
	second, _ := counter.add("192.0.2.10")
	_, emptyAdded := counter.add("")

	fmt.Println("first add: ", first, added)
	fmt.Println("second add: ", second)
	fmt.Println("empty IP added:", emptyAdded)

	count, exists := counter.count("192.0.2.10")
	fmt.Println("known IP: ", count, exists)

	count, exists = counter.count("203.0.113.7")
	fmt.Println("missing IP:", count, exists)

	counter.remove("192.0.2.10")
	_, exists = counter.count("192.0.2.10")
	fmt.Println("exists after remove: ", exists)
}