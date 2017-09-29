package output

import (
	"time"

	"github.com/theothertomelliott/uilive"
	"github.com/nedscode/nedward/tracker"
)

type Follower struct {
	inProgress *InProgressRenderer
	writer     *uilive.Writer
}

func NewFollower() *Follower {
	uilive.RefreshInterval = time.Hour
	f := &Follower{
		inProgress: NewInProgressRenderer(),
	}
	f.Reset()
	return f
}

func (f *Follower) Reset() {
	if f.writer != nil {
		panic("Follower not stopped correctly")
	}
	f.writer = uilive.New()
	f.writer.Start()
}

func (f *Follower) Handle(update tracker.Task) {
	state := update.State()
	if state != tracker.TaskStatePending &&
		state != tracker.TaskStateInProgress {
		bp := f.writer.Bypass()
		renderer := NewCompletionRenderer(update)
		renderer.Render(bp)
	}

	f.inProgress.Render(f.writer, update)
	f.writer.Flush()
}

func (f *Follower) Done() {
	f.writer.Stop()
	f.writer = nil
}
