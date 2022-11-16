package tx_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/stretchr/testify/suite"
)

// TODO: implement this latter
type IntegrationTestSuite struct {
	suite.Suite

	cfg         network.Config
	network     *network.Network
	txHeight    int64
	queryClient tx.ServiceClient
	txRes       sdk.TxResponse
}

func (s *IntegrationTestSuite) SetupTest() {
	fmt.Println("SetupTest")
}

func (s *IntegrationTestSuite) TearDownTest() {
	fmt.Println("TearDownTest")
}

func TestIntergartionTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
