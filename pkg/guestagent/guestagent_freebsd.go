package guestagent

import (
	"context"
	"os"
	"reflect"
	"time"

	"github.com/lima-vm/lima/pkg/guestagent/api"
	"github.com/lima-vm/lima/pkg/guestagent/sockstat"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func New(newTicker func() (<-chan time.Time, func()), _ time.Duration) (Agent, error) {
	a := &agent{
		newTicker: newTicker,
	}

	return a, nil
}

type agent struct {
	// Ticker is like time.Ticker.
	newTicker func() (<-chan time.Time, func())
}

type eventState struct {
	ports []*api.IPPort
}

func comparePorts(old, neww []*api.IPPort) (added, removed []*api.IPPort) {
	mRaw := make(map[string]*api.IPPort, len(old))
	mStillExist := make(map[string]bool, len(old))

	for _, f := range old {
		k := f.String()
		mRaw[k] = f
		mStillExist[k] = false
	}
	for _, f := range neww {
		k := f.String()
		if _, ok := mRaw[k]; !ok {
			added = append(added, f)
		}
		mStillExist[k] = true
	}

	for k, stillExist := range mStillExist {
		if !stillExist {
			if x, ok := mRaw[k]; ok {
				removed = append(removed, x)
			}
		}
	}
	return
}

func (a *agent) collectEvent(ctx context.Context, st eventState) (*api.Event, eventState) {
	var (
		ev  = &api.Event{}
		err error
	)
	newSt := st
	newSt.ports, err = a.LocalPorts(ctx)
	if err != nil {
		ev.Errors = append(ev.Errors, err.Error())
		ev.Time = timestamppb.Now()
		return ev, newSt
	}
	ev.LocalPortsAdded, ev.LocalPortsRemoved = comparePorts(st.ports, newSt.ports)
	ev.Time = timestamppb.Now()
	return ev, newSt
}

func isEventEmpty(ev *api.Event) bool {
	empty := &api.Event{}
	copied := ev
	copied.Time = nil
	return reflect.DeepEqual(empty, copied)
}

func (a *agent) Events(ctx context.Context, ch chan *api.Event) {
	defer close(ch)
	tickerCh, tickerClose := a.newTicker()
	defer tickerClose()
	var st eventState
	for {
		var ev *api.Event
		ev, st = a.collectEvent(ctx, st)
		if !isEventEmpty(ev) {
			ch <- ev
		}
		select {
		case <-ctx.Done():
			return
		case _, ok := <-tickerCh:
			if !ok {
				return
			}
			logrus.Debug("tick!")
		}
	}
}

func (a *agent) LocalPorts(_ context.Context) ([]*api.IPPort, error) {
	var res []*api.IPPort
	tcpParsed, err := sockstat.ParseOutput()
	if err != nil {
		return res, err
	}

	for _, f := range tcpParsed {
		switch f.Kind {
		case sockstat.TCP4, sockstat.TCP6:
		default:
			continue
		}
		if f.State == sockstat.Listen {
			res = append(res,
				&api.IPPort{
					Ip:   f.IP.String(),
					Port: int32(f.Port),
					Protocol: "tcp",
				})
		}
	}

	return res, nil
}

func (a *agent) Info(ctx context.Context) (*api.Info, error) {
	var (
		info api.Info
		err  error
	)
	info.LocalPorts, err = a.LocalPorts(ctx)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (a *agent) HandleInotify(event *api.Inotify) {
	location := event.MountPath
	if _, err := os.Stat(location); err == nil {
		local := event.Time.AsTime().Local()
		err := os.Chtimes(location, local, local)
		if err != nil {
			logrus.Errorf("error in inotify handle. Event: %s, Error: %s", event, err)
		}
	}
}
