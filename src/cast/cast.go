package cast

type Channel interface {
	Join(chan <- interface{})
	Leave(chan <- interface{})
	Send(interface{})
	Close()
}

type caster struct {
	in      chan interface{}
	join    chan chan <- interface{}
	leave   chan chan <- interface{}
	members map[chan <- interface{}]bool
}

func New() *caster {
	c := &caster{
		in: make(chan interface{}),
		join: make(chan chan <- interface{}),
		leave: make(chan chan <- interface{}),
		members: make(map[chan <- interface{}]bool),
	}
	go c.run()
	return c
}

func (this *caster) broadcast(data interface{}) {
	for c := range this.members {
		c <- data
	}
}

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

func (this *caster) Join(ch chan <- interface{}) {
	this.join <- ch
}

func (this *caster) Leave(ch chan <- interface{}) {
	this.leave <- ch
}

func (this *caster) Close() {
	close(this.join)
}

func (this *caster) Send(data interface{}) {
	this.in <- data
}
