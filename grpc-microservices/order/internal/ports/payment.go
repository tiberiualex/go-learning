package ports

import "github.com/tiberiualex/go-learning/grpc-microservices/order/internal/application/core/domain"

type PaymentPort interface {
	Charge(*domain.Order) error
}
