package context

import (
	"github.com/encobrain/go-context.v2"
	"strconv"
)

type deadlockState struct {
	childs map[string]*deadlockState
	count  int
}

func newDeadlockState() *deadlockState {
	return &deadlockState{
		childs: map[string]*deadlockState{},
	}
}

func getDeadlockState(ctx context.Context, state *deadlockState) {
	childs := ctx.Childs()

	for _, ch := range childs {
		name := "▶"

		select {
		case <-ch.Finished(false):
			name = "■"
		default:
		}

		name += " " + ch.Name()

		st := state.childs[name]

		if st == nil {
			st = newDeadlockState()
			state.childs[name] = st
		}

		st.count++

		getDeadlockState(ch, st)
	}
}

func getDeadlockInfo(st *deadlockState, level string) (info string) {
	for name, inf := range st.childs {
		info += level + name + " x" + strconv.Itoa(inf.count) + "\n"
		info += getDeadlockInfo(inf, level+"   ")
	}

	return info
}

func GetDeadlocksInfo(ctx context.Context) (info string) {
	st := newDeadlockState()

	getDeadlockState(ctx, st)

	return getDeadlockInfo(st, "")
}
