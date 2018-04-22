package img

type Queue struct {
	ops chan *Command
}

type OpCallback func()

func NewQueue() *Queue {
	q := &Queue{}
	q.ops = make(chan *Command)
	go q.start()
	return q
}

func (q *Queue) start() {
	for op := range q.ops {
		if op.Result == nil {
			if op.Resize != nil {
				op.Result, op.Err = op.Resize(op.Image, op.Size, op.ImgId, op.SupportedFormats)
			} else if op.Optimise != nil {
				op.Result, op.Err = op.Optimise(op.Image, op.ImgId, op.SupportedFormats)
			}
		}
		op.Finished = true
		op.FinishedCond.Signal()
	}
}

func (q *Queue) AddAndWait(op *Command, callback OpCallback) {
	//Adding operation to the execution channel
	q.ops <- op

	//Waiting for operation to finish
	op.FinishedCond.L.Lock()
	for !op.Finished {
		op.FinishedCond.Wait()
	}
	op.FinishedCond.L.Unlock()

	callback()
}
