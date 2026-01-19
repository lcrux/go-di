package main

import (
	"fmt"
	"go-di/demo/models"
	"go-di/demo/repositories"
	"go-di/demo/services"
	"log"

	"github.com/lcrux/go-di/registry"
)

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
	helloSvc   HelloService
	anotherSvc AnotherService
	message    string
}

func (d *DummyServiceImpl) DoSomething() {
	d.helloSvc.SayHello(d.message)
}

func NewHelloService() HelloService {
	return &HelloServiceImpl{}
}

type HelloService interface {
	SayHello(name string)
}

type HelloServiceImpl struct {
}

func (h *HelloServiceImpl) SayHello(name string) {
	fmt.Printf("Hello, %s!\n", name)
}

type AnotherService interface {
	DoAnotherThing()
}

func init() {
	registry.Register[services.OrderService](services.NewOrderService, registry.Transient)
	registry.Register[services.UserService](services.NewUserService, registry.Scoped)
	registry.Register[repositories.OrderRepository](repositories.NewOrderRepository, registry.Singleton)
	registry.Register[repositories.UserRepository](repositories.NewUserRepository, registry.Singleton)
	registry.Register[DummyService](func(helloSvc HelloService) DummyService {
		return NewDummyService(helloSvc, "I am a DummyService")
	}, registry.Transient)
	registry.Register[HelloService](NewHelloService, registry.Transient)
}

func main() {
	dummyService, err := registry.Resolve[DummyService]()
	if err != nil {
		log.Printf("Failed to resolve DummyService: %v\n", err)
		return
	}
	dummyService.DoSomething()

	orderService, err := registry.Resolve[services.OrderService]()
	if err != nil {
		log.Printf("Failed to resolve OrderService: %v\n", err)
		return
	}

	userService, err := registry.Resolve[services.UserService]()
	if err != nil {
		log.Printf("Failed to resolve UserService: %v\n", err)
		return
	}

	createdOrder, err := orderService.CreateOrder(&models.Order{
		UserID:      1,
		ProductName: "Product 1",
		Amount:      100.0,
	})
	if err != nil {
		log.Printf("Failed to create order: %v\n", err)
		return
	}
	log.Printf("Created order: %+v\n", createdOrder)

	users, err := userService.GetUsers(&models.Pagination{Page: 1, PageSize: 5})
	if err != nil {
		log.Printf("Failed to get users: %v\n", err)
		return
	}
	log.Printf("Retrieved users: %+v\n", users)

	user, err := userService.GetUser(1)
	if err != nil {
		log.Printf("Failed to get user: %v\n", err)
		return
	}
	log.Printf("Retrieved user: %+v\n", user)

	log.Println("Dependency Injection Container in Go")

	regCtx := registry.NewRegistryContext()
	defer regCtx.Close()

	usrService, err := registry.ResolveWithContext[services.UserService](regCtx)
	if err != nil {
		log.Printf("Failed to resolve UserService with context: %v\n", err)
		return
	}

	usr, err := usrService.GetUser(1)
	if err != nil {
		log.Printf("Failed to get user with context: %v\n", err)
		return
	}
	log.Printf("Retrieved user with context: %+v\n", usr)
}
