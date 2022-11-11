package playbook

import "sync"

var gawait sync.Map

func Await(ss []string, ignore bool) error {
	for _, s := range ss {
		v, ok := gawait.Load(s)
		if !ok {
			continue
		}
		err := <-v.(chan error)

		if err != nil && !ignore {
			return err
		}

	}
	return nil
}

func AddAwait(s string) {
	a := make(chan error)
	gawait.Store(s, a)
}

func DoneAwait(s string, err error) {
	v, ok := gawait.Load(s)
	if !ok {
		return
	}
	go func() {
		v.(chan error) <- err
	}()
}
