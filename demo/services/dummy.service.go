package services

import "log"

func NewDummyService(helloSvc HelloService, message string) DummyService {
	if helloSvc == nil {
		log.Fatal("HelloService cannot be nil")
	}
	return &DummyServiceImpl{
		message:  message,
		helloSvc: helloSvc,
	}
}

type DummyService interface {
	DoSomething()
}

type DummyServiceImpl struct {
	helloSvc HelloService
	message  string
}

func (d *DummyServiceImpl) DoSomething() {
	d.helloSvc.SayHello(d.message)
}
