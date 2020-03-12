package mock

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/epochStart"
)

// EpochStartNotifierStub -
type EpochStartNotifierStub struct {
	RegisterHandlerCalled   func(handler epochStart.EpochStartHandler)
	UnregisterHandlerCalled func(handler epochStart.EpochStartHandler)
	NotifyAllCalled         func(hdr data.HeaderHandler)
	NotifyAllPrepareCalled  func(hdr data.HeaderHandler)
	epochStartHdls          []epochStart.EpochStartHandler
}

// RegisterHandler -
func (esnm *EpochStartNotifierStub) RegisterHandler(handler epochStart.EpochStartHandler) {
	if esnm.RegisterHandlerCalled != nil {
		esnm.RegisterHandlerCalled(handler)
	}

	esnm.epochStartHdls = append(esnm.epochStartHdls, handler)
}

// UnregisterHandler -
func (esnm *EpochStartNotifierStub) UnregisterHandler(handler epochStart.EpochStartHandler) {
	if esnm.UnregisterHandlerCalled != nil {
		esnm.UnregisterHandlerCalled(handler)
	}

	for i, hdl := range esnm.epochStartHdls {
		if hdl == handler {
			esnm.epochStartHdls = append(esnm.epochStartHdls[:i], esnm.epochStartHdls[i+1:]...)
			break
		}
	}
}

// NotifyAllPrepare -
func (esnm *EpochStartNotifierStub) NotifyAllPrepare(metaHeader data.HeaderHandler) {
	if esnm.NotifyAllPrepareCalled != nil {
		esnm.NotifyAllPrepareCalled(metaHeader)
	}

	for _, hdl := range esnm.epochStartHdls {
		hdl.EpochStartPrepare(metaHeader)
	}
}

// NotifyAll -
func (esnm *EpochStartNotifierStub) NotifyAll(hdr data.HeaderHandler) {
	if esnm.NotifyAllCalled != nil {
		esnm.NotifyAllCalled(hdr)
	}

	for _, hdl := range esnm.epochStartHdls {
		hdl.EpochStartAction(hdr)
	}
}

// IsInterfaceNil -
func (esnm *EpochStartNotifierStub) IsInterfaceNil() bool {
	return esnm == nil
}
