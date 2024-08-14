package customer

import (
	"context"
	"log"
	"testing"

	"testcontainers-go-demo/testhelpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CustomerRepoIsolatedTestSuite struct {
	suite.Suite
	ctx context.Context
}

func (suite *CustomerRepoIsolatedTestSuite) SetupSuite() {
	suite.ctx = context.Background()
}

func (suite *CustomerRepoIsolatedTestSuite) GetRepository() (*Repository, error) {
	t := suite.T()
	pgContainer, err := testhelpers.CreatePostgresContainer(suite.ctx)
	if err != nil {
		return nil, err
	}

	repository, err := NewRepository(suite.ctx, pgContainer.ConnectionString)
	if err != nil {
		return nil, err
	}

	t.Cleanup(func() {
		if err := pgContainer.Terminate(suite.ctx); err != nil {
			t.Fatalf("failed to terminate pgContainer: %s", err)
		}
	})

	return repository, nil
}

func (suite *CustomerRepoIsolatedTestSuite) TestCreateCustomer() {
	t := suite.T()
	t.Parallel()

	repository, err := suite.GetRepository()
	if err != nil {
		log.Fatal(err)
	}

	log.Println(repository.conn.Config().ConnString())

	customer, err := repository.CreateCustomer(suite.ctx, Customer{
		Name:  "Frank",
		Email: "frank@gmail.com",
	})
	assert.NoError(t, err)
	assert.NotNil(t, customer.Id)
}

func (suite *CustomerRepoIsolatedTestSuite) TestGetCustomerByEmail() {
	t := suite.T()
	t.Parallel()

	repository, err := suite.GetRepository()
	if err != nil {
		log.Fatal(err)
	}

	log.Println(repository.conn.Config().ConnString())

	customer, err := repository.GetCustomerByEmail(suite.ctx, "beth@gmail.com")
	assert.NoError(t, err)
	assert.NotNil(t, customer)
	assert.Equal(t, "Beth", customer.Name)
	assert.Equal(t, "beth@gmail.com", customer.Email)
}

func (suite *CustomerRepoIsolatedTestSuite) TestDeleteCustomerByEmailIsolated() {
	t := suite.T()
	t.Parallel()

	repository, err := suite.GetRepository()
	if err != nil {
		log.Fatal(err)
	}

	log.Println(repository.conn.Config().ConnString())

	err = repository.DeleteCustomerByEmail(suite.ctx, "beth@gmail.com")
	assert.NoError(t, err)

	customer, err := repository.GetCustomerByEmail(suite.ctx, "beth@gmail.com")
	assert.Error(t, err)
	assert.ErrorContains(t, err, "no rows in result set")
	assert.Empty(t, customer.Email)
}

func TestCustomerRepoIsolatedTestSuite(t *testing.T) {
	suite.Run(t, new(CustomerRepoIsolatedTestSuite))
}
