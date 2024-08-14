package customer

import (
	"context"
	"log"
	"os"
	"testing"

	"testcontainers-go-demo/testhelpers"

	"github.com/stretchr/testify/assert"
)

func GetRepository(t *testing.T, ctx context.Context) (*Repository, error) {
	//variable must be set when using rancher
	os.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", "/var/run/docker.sock")
	pgContainer, err := testhelpers.CreatePostgresContainer(ctx)
	if err != nil {
		return nil, err
	}

	repository, err := NewRepository(ctx, pgContainer.ConnectionString)
	if err != nil {
		return nil, err
	}

	t.Cleanup(func() {
		log.Printf("removing container with id %s", pgContainer.GetContainerID())
		if err := pgContainer.Terminate(ctx); err != nil {
			log.Fatalf("error terminating postgres container: %s", err)
		}
	})

	return repository, nil
}

func TestCreateCustomerIsolated(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	repository, err := GetRepository(t, ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(repository.conn.Config().ConnString())

	customer, err := repository.CreateCustomer(ctx, Customer{
		Name:  "Allison",
		Email: "allison@gmail.com",
	})
	assert.NoError(t, err)
	assert.NotNil(t, customer.Id)
}

func TestGetCustomerByEmailIsolated(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	repository, err := GetRepository(t, ctx)
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

func TestDeleteCustomerByEmailIsolated(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	repository, err := GetRepository(t, ctx)
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
