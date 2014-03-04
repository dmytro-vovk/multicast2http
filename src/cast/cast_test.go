package cast

import (
	"testing"
	"sync"
)

const msg int = 17

func TestCast(t *testing.T) {
	wgBefore := sync.WaitGroup{}
	wgAfter := sync.WaitGroup{}
	c := New()
	defer c.Close()
	for i := 1; i < 6; i++ {
		wgBefore.Add(1)
		go client(c, t, &wgBefore, &wgAfter)
	}
	// Give some time for goroutines to settle
	wgBefore.Wait()
	c.Send(msg)
	wgAfter.Wait()
}

func client(c *caster, t *testing.T, wg1, wg2 *sync.WaitGroup) {
	ch := make(chan interface{})
	c.Join(ch)
	wg1.Done()
	defer c.Leave(ch)
	defer wg2.Done()
	select {
	case r := <-ch:
		if r == msg {
			return
		} else {
			t.Errorf("Got wrong broadcast message")
		}
	}
	t.Error("Did not get anything :(")
}
