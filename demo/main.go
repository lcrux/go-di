package main

import (
	"go-di/demo/models"
	"go-di/demo/repositories"
	"go-di/demo/services"
	"log"

	godi "github.com/lcrux/go-di"
)

func init() {
	godi.Register[services.OrderService](services.NewOrderService, godi.Transient)
	godi.Register[services.UserService](services.NewUserService, godi.Scoped)
	godi.Register[repositories.OrderRepository](repositories.NewOrderRepository, godi.Singleton)
	godi.Register[repositories.UserRepository](repositories.NewUserRepository, godi.Singleton)

	godi.Register[services.DummyService](func(helloSvc services.HelloService) services.DummyService {
		return services.NewDummyService(helloSvc, "I am a DummyService")
	}, godi.Transient)
	godi.Register[services.HelloService](services.NewHelloService, godi.Transient)
}

func main() {
	dummyService, err := godi.Resolve[services.DummyService]()
	if err != nil {
		log.Printf("Failed to resolve DummyService: %v\n", err)
		return
	}
	dummyService.DoSomething()

	orderService, err := godi.Resolve[services.OrderService]()
	if err != nil {
		log.Printf("Failed to resolve OrderService: %v\n", err)
		return
	}

	userService, err := godi.Resolve[services.UserService]()
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

	regCtx := godi.NewRegistryContext()
	defer regCtx.Close()

	usrService, err := godi.ResolveWithContext[services.UserService](regCtx)
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

	log.Println("Dependency Injection Container in Go")
}
