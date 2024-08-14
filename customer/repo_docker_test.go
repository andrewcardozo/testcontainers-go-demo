package customer

import (
	"context"
	"log"

	"testing"

	"testcontainers-go-demo/testhelpers"

	"github.com/stretchr/testify/assert"
)

func TestCustomerRepositoryDocker(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := testhelpers.CreatePostgresDockerContainer(ctx, t)
	if err != nil {
		log.Fatal(err)
	}

	customerRepo, err := NewRepository(ctx, pgContainer.ConnectionString)
	assert.NoError(t, err)

	c, err := customerRepo.CreateCustomer(ctx, Customer{
		Name:  "Tom",
		Email: "tom@gmail.com",
	})
	assert.NoError(t, err)
	assert.NotNil(t, c)

	customer, err := customerRepo.GetCustomerByEmail(ctx, "tom@gmail.com")
	assert.NoError(t, err)
	assert.NotNil(t, customer)
	assert.Equal(t, "Tom", customer.Name)
	assert.Equal(t, "tom@gmail.com", customer.Email)
}
