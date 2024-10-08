package repository

import (
	"context"
	"database/sql"
	"simple-restaurant-web/helper"
	"simple-restaurant-web/model/domain"
)

type OrderRepositoryImpl struct {
}

func NewOrderRepository() OrderRepository {
	return &OrderRepositoryImpl{}
}

func (repository *OrderRepositoryImpl) Save(ctx context.Context, Tx *sql.Tx, order domain.Orders) domain.Orders {
	var idOrder int
	order.IdCustomer = ctx.Value("idCustomer").(int)
	SQL := `INSERT INTO "order"(id_customer) VALUES($1) RETURNING id`
	err := Tx.QueryRowContext(ctx, SQL, order.IdCustomer).Scan(&idOrder)
	helper.PanicIfError(err)
	order.Id = idOrder

	for _, detail := range order.OrderDetails {
		SQL = "INSERT INTO order_detail(order_id, food_id, quantity) VALUES($1, $2, $3)"
		_, err := Tx.ExecContext(ctx, SQL, idOrder, detail.FoodId, detail.Quantity)
		helper.PanicIfError(err)
	}

	SQL = "SELECT SUM(price * quantity) as total_price, SUM(quantity) as total_quantity FROM order_detail JOIN food ON food.id = order_detail.food_id WHERE order_id = $1"
	err = Tx.QueryRowContext(ctx, SQL, idOrder).Scan(&order.TotalPrice, &order.Quantity)
	helper.PanicIfError(err)

	SQL = `UPDATE "order" SET total_quantity = $1, total_price = $2 WHERE id = $3`
	_, err = Tx.ExecContext(ctx, SQL, order.Quantity, order.TotalPrice, idOrder)
	helper.PanicIfError(err)

	order.OrderDetails = []domain.OrderDetail{}

	SQL = "SELECT name, price, quantity FROM order_detail JOIN food ON food.id = order_detail.food_id WHERE order_id = $1"
	rows, err := Tx.QueryContext(ctx, SQL, idOrder)
	helper.PanicIfError(err)
	defer rows.Close()

	for rows.Next() {
		newOrderDetail := domain.OrderDetail{}
		err := rows.Scan(&newOrderDetail.FoodName, &newOrderDetail.FoodPrice, &newOrderDetail.Quantity)
		helper.PanicIfError(err)
		order.OrderDetails = append(order.OrderDetails, newOrderDetail)
	}

	return order
}

func (repository *OrderRepositoryImpl) Get(ctx context.Context, Tx *sql.Tx) []domain.Orders {
	SQL := `SELECT id, total_quantity, total_price, id_customer FROM "order" WHERE id_customer = $1`
	rows, err := Tx.QueryContext(ctx, SQL, ctx.Value("idCustomer"))
	helper.PanicIfError(err)
	defer rows.Close()

	orders := []domain.Orders{}
	for rows.Next() {
		order := domain.Orders{}
		err := rows.Scan(&order.Id, &order.Quantity, &order.TotalPrice, &order.IdCustomer)
		helper.PanicIfError(err)
		orders = append(orders, order)
	}

	return orders
}

func (repository *OrderRepositoryImpl) GetDetail(ctx context.Context, Tx *sql.Tx, orderId int) domain.Orders {
	var order domain.Orders
	SQL := `SELECT id, total_quantity, total_price FROM "order" WHERE id_customer = $1 AND id = $2`
	err := Tx.QueryRowContext(ctx, SQL, ctx.Value("idCustomer"), orderId).Scan(&order.Id, &order.Quantity, &order.TotalPrice)
	helper.PanicIfError(err)

	order.OrderDetails = []domain.OrderDetail{}

	SQL = "SELECT food.name, food.price, order_detail.quantity FROM order_detail JOIN food ON food.id = order_detail.food_id WHERE order_id = $1"
	rows, err := Tx.QueryContext(ctx, SQL, order.Id)
	helper.PanicIfError(err)
	defer rows.Close()

	for rows.Next() {
		orderDetail := domain.OrderDetail{}
		err := rows.Scan(&orderDetail.FoodName, &orderDetail.FoodPrice, &orderDetail.Quantity)
		helper.PanicIfError(err)
		order.OrderDetails = append(order.OrderDetails, orderDetail)
	}

	return order
}
