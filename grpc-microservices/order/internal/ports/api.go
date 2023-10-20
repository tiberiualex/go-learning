package ports

import "github.com/tiberiualex/go-learning/grpc-microservices/order/internal/application/core/domain"

type APIPort interface {
	PlaceOrder(order domain.Order) (domain.Order, error)
}
