package ginmetrics

import "sync"

type PararelQuery struct {
	sync.Mutex
	w   sync.WaitGroup
	Err []error
}

func NewPararelAction() *PararelQuery {
	return &PararelQuery{
		Mutex: sync.Mutex{},
		Err:   []error{},
		w:     sync.WaitGroup{},
	}
}

func (p *PararelQuery) Wait() error {
	p.w.Wait()

	if len(p.Err) == 0 {
		return nil
	}

	return p.Err[0]
}

func (p *PararelQuery) SetError(err error) {
	if err == nil {
		return
	}

	p.Lock()
	defer p.Unlock()

	p.Err = append(p.Err, err)

}

func (p *PararelQuery) Add(handle func() error) {
	p.w.Add(1)
	go func() {
		defer p.w.Done()
		err := handle()
		p.SetError(err)
	}()
}
