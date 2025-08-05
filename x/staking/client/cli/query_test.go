package cli_test

import (
	"fmt"
	"io"
	"testing"

	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/client/cli"
)

type QueryTestSuite struct {
	suite.Suite

	kr      keyring.Keyring
	encCfg  testutilmod.TestEncodingConfig
	baseCtx client.Context
}

func TestQueryTestSuite(t *testing.T) {
	suite.Run(t, new(QueryTestSuite))
}

func (s *QueryTestSuite) SetupSuite() {
	s.encCfg = testutilmod.MakeTestEncodingConfig(staking.AppModuleBasic{})
	s.kr = keyring.NewInMemory(s.encCfg.Codec)
	s.baseCtx = client.Context{}.
		WithKeyring(s.kr).
		WithTxConfig(s.encCfg.TxConfig).
		WithCodec(s.encCfg.Codec).
		WithClient(clitestutil.MockCometRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard).
		WithChainID("test-chain")
}

func (s *QueryTestSuite) TestCmdQueryProposers() {
	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expectedErrMsg string
	}{
		{
			name:      "valid query",
			args:      []string{fmt.Sprintf("--%s=json", flags.FlagOutput)},
			expectErr: false,
		},
		{
			name:      "valid query with height",
			args:      []string{fmt.Sprintf("--%s=1", flags.FlagHeight), fmt.Sprintf("--%s=json", flags.FlagOutput)},
			expectErr: false,
		},
		{
			name:           "invalid flag",
			args:           []string{"--invalid-flag"},
			expectErr:      true,
			expectedErrMsg: "unknown flag",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.CmdQueryProposers()
			clientCtx := s.baseCtx

			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
				if tc.expectedErrMsg != "" {
					s.Require().Contains(err.Error(), tc.expectedErrMsg)
				}
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(out)
			}
		})
	}
}
