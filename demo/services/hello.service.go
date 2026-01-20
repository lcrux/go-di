package services

import "log"

func NewHelloService(hellosvc HelloService) HelloService {
	return &HelloServiceImpl{}
}

type HelloService interface {
	SayHello(name string)
}

type HelloServiceImpl struct {
}

func (h *HelloServiceImpl) SayHello(name string) {
	log.Printf("Hello, %s!\n", name)
}
