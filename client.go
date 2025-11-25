package polymarket

import (
	"context"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ivanzzeth/ethclient"
	polymarketclob "github.com/ivanzzeth/polymarket-go-clob-client"
	clobconst "github.com/ivanzzeth/polymarket-go-clob-client/constants"

	polymarketcontracts "github.com/ivanzzeth/polymarket-go-contracts"
	polymarketdata "github.com/ivanzzeth/polymarket-go-data-client"
	polymarketgamma "github.com/ivanzzeth/polymarket-go-gamma-client"
	polymarketrealtime "github.com/ivanzzeth/polymarket-go-real-time-data-client"
)

type Client struct {
	gammaClient        *polymarketgamma.Client
	dataClient         *polymarketdata.Client
	realtimeDataClient *polymarketrealtime.Client
	contractInterface  *polymarketcontracts.ContractInterface
	clobClient         *polymarketclob.Client
}

type ClientConfig struct {
	RealtimeDataClientOptions []polymarketrealtime.ClientOption
	ContractInterfaceOptions  []polymarketcontracts.ContractInterfaceOption
	ClobClientOptions         []polymarketclob.ClobClientOption
}

type ClientOption func(c *ClientConfig)

func WithRealTimeOptions(options ...polymarketrealtime.ClientOption) ClientOption {
	return func(c *ClientConfig) {
		c.RealtimeDataClientOptions = options
	}
}

func WithContractInterfaceOptions(options ...polymarketcontracts.ContractInterfaceOption) ClientOption {
	return func(c *ClientConfig) {
		c.ContractInterfaceOptions = options
	}
}

func WithClobClientOptions(options ...polymarketclob.ClobClientOption) ClientOption {
	return func(c *ClientConfig) {
		c.ClobClientOptions = options
	}
}

func NewClient(ethclient ethclient.EthClientInterface, options ...ClientOption) (*Client, error) {
	defaultOptions := &ClientConfig{}

	for _, opFn := range options {
		opFn(defaultOptions)
	}

	chainID, err := ethclient.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	httpClient := http.DefaultClient

	gammaClient := polymarketgamma.NewClient(httpClient)
	dataClient, err := polymarketdata.NewClient(httpClient)
	if err != nil {
		return nil, err
	}

	realtimeDataClient := polymarketrealtime.New(defaultOptions.RealtimeDataClientOptions...)
	contractInterface, err := polymarketcontracts.NewContractInterface(ethclient, defaultOptions.ContractInterfaceOptions...)
	if err != nil {
		return nil, err
	}

	clobClient, err := polymarketclob.NewClient(clobconst.CLOB_API_URL, chainID.Int64(), defaultOptions.ClobClientOptions...)
	if err != nil {
		return nil, err
	}

	return &Client{
		gammaClient:        gammaClient,
		dataClient:         dataClient,
		realtimeDataClient: realtimeDataClient,
		contractInterface:  contractInterface,
		clobClient:         clobClient,
	}, nil
}

// GammaClient returns the gamma client
func (c *Client) GammaClient() *polymarketgamma.Client {
	return c.gammaClient
}

// DataClient returns the data client
func (c *Client) DataClient() *polymarketdata.Client {
	return c.dataClient
}

// RealtimeDataClient returns the realtime data client
func (c *Client) RealtimeDataClient() *polymarketrealtime.Client {
	return c.realtimeDataClient
}

// ContractInterface returns the contract interface
func (c *Client) ContractInterface() *polymarketcontracts.ContractInterface {
	return c.contractInterface
}

// ClobClient returns the CLOB client
func (c *Client) ClobClient() *polymarketclob.Client {
	return c.clobClient
}

func (c *Client) EnableTrading(ctx context.Context) ([]common.Hash, error) {
	return c.contractInterface.EnableTrading(ctx)
}

// Split splits collateral into conditional tokens for a binary market
// Uses standard Polymarket partition [1, 2] for binary outcomes (Yes/No)
func (c *Client) Split(ctx context.Context, conditionId common.Hash, amount *big.Int) (common.Hash, error) {
	partition := []*big.Int{big.NewInt(1), big.NewInt(2)}
	return c.contractInterface.Split(ctx, conditionId, partition, amount)
}

// Merge merges conditional tokens back into collateral for a binary market
// Uses standard Polymarket partition [1, 2] for binary outcomes (Yes/No)
func (c *Client) Merge(ctx context.Context, conditionId common.Hash, amount *big.Int) (common.Hash, error) {
	partition := []*big.Int{big.NewInt(1), big.NewInt(2)}
	return c.contractInterface.Merge(ctx, conditionId, partition, amount)
}

// Redeem redeems conditional tokens for a resolved binary market
// Uses standard Polymarket indexSets [1, 2] for binary outcomes (Yes/No)
func (c *Client) Redeem(ctx context.Context, conditionId common.Hash) (common.Hash, error) {
	indexSets := []*big.Int{big.NewInt(1), big.NewInt(2)}
	return c.contractInterface.Redeem(ctx, conditionId, indexSets)
}

// RedeemNegRisk redeems NegRisk market positions
// amounts is a slice containing the amount to redeem for each outcome
// For binary NegRisk markets, use a slice of two amounts [yesAmount, noAmount]
func (c *Client) RedeemNegRisk(ctx context.Context, conditionId common.Hash, amounts []*big.Int) (common.Hash, error) {
	return c.contractInterface.RedeemNegRisk(ctx, conditionId, amounts)
}

// DeploySafe deploys a Gnosis Safe wallet for the configured signer
// Returns the Safe proxy address and the deployment transaction hash
func (c *Client) DeploySafe() (safeProxy common.Address, txHash common.Hash, err error) {
	return c.contractInterface.DeploySafe()
}
