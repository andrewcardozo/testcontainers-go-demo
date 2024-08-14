package customer

import (
	"context"
	"log"
	"testing"

	"testcontainers-go-demo/testhelpers"

	"github.com/stretchr/testify/assert"
)

func GetDockerRepository(t *testing.T, ctx context.Context) (*Repository, error) {
	pgContainer, err := testhelpers.CreatePostgresDockerContainer(ctx, t)
	if err != nil {
		return nil, err
	}

	repository, err := NewRepository(ctx, pgContainer.ConnectionString)
	if err != nil {
		return nil, err
	}

	return repository, nil
}

func TestCreateCustomerDockerIsolated(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	repository, err := GetDockerRepository(t, ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(repository.conn.Config().ConnString())

	customer, err := repository.CreateCustomer(ctx, Customer{
		Name:  "Kyle",
		Email: "kyle@gmail.com",
	})
	assert.NoError(t, err)
	assert.NotNil(t, customer.Id)
}

func TestGetCustomerByEmailDockerIsolated(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	repository, err := GetDockerRepository(t, ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(repository.conn.Config().ConnString())
	customer, err := repository.GetCustomerByEmail(ctx, "phil@gmail.com")
	assert.NoError(t, err)
	assert.NotNil(t, customer)
	assert.Equal(t, "Phil", customer.Name)
	assert.Equal(t, "phil@gmail.com", customer.Email)
}

func TestDeleteCustomerByEmailDockerIsolated(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	repository, err := GetDockerRepository(t, ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(repository.conn.Config().ConnString())
	err = repository.DeleteCustomerByEmail(ctx, "phil@gmail.com")
	assert.NoError(t, err)

	customer, err := repository.GetCustomerByEmail(ctx, "phil@gmail.com")
	assert.Error(t, err)
	assert.ErrorContains(t, err, "no rows in result set")
	assert.Empty(t, customer.Email)
}
