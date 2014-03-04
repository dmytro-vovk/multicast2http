package cast

type Channel interface {
	Join(chan <- interface{})
	Leave(chan <- interface{})
	Send(interface{})
	Close()
}

type caster struct {
	in      chan interface{}             // Messages here
	join    chan chan <- interface{}     // Knock into this chan to get subscribed
	leave   chan chan <- interface{}     // ...and unsubscribed
	members map[chan <- interface{}]bool // List of active listeners
}

// Return new broadcaster instance
func New(queueLength uint) *caster {
	c := &caster{
		in: make(chan interface{}, queueLength),
		join: make(chan chan <- interface{}),
		leave: make(chan chan <- interface{}),
		members: make(map[chan <- interface{}]bool),
	}
	go c.run()
	return c
}

// Do sending data to all listeners
func (this *caster) broadcast(data interface{}) {
	for c := range this.members {
		// Make sending data to members non-blocking
		go func(ch chan <- interface{}, m interface{}) {
			ch <- m
		}(c, data)
	}
}

// Process all messages
func (this *caster) run() {
	for {
		select {
		case data := <-this.in:
			this.broadcast(data)
		case c, ok := <-this.join:
			if ok {
				this.members[c] = true
			} else {
				return
			}
		case c := <-this.leave:
			delete(this.members, c)
		}
	}
}

// Members supply their own channel they will listen on
func (this *caster) Join(ch chan <- interface{}) {
	this.join <- ch
}

// Member wants to leave the broadcast
func (this *caster) Leave(ch chan <- interface{}) {
	this.leave <- ch
}

// Close the broadcast channel
func (this *caster) Close() {
	close(this.join)
}

// Send data to all listeners
func (this *caster) Send(data interface{}) {
	this.in <- data
}
