package keeper_test

import (
	"encoding/base64"
	"math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Pylons-tech/pylons/x/pylons/keeper"
	"github.com/Pylons-tech/pylons/x/pylons/types"
)

func (suite *IntegrationTestSuite) TestFulfillTradeMsgServerSimple() {
	k := suite.k
	ctx := suite.ctx
	require := suite.Require()

	wctx := sdk.WrapSDKContext(ctx)
	srv := keeper.NewMsgServerImpl(k)

	creatorA := types.GenTestBech32FromString("creatorA")
	creatorB := types.GenTestBech32FromString("creatorB")

	for i := 0; i < 5; i++ {
		msgCreate := &types.MsgCreateTrade{
			Creator:     creatorA,
			CoinInputs:  nil,
			ItemInputs:  nil,
			CoinOutputs: nil,
			ItemOutputs: nil,
			ExtraInfo:   "extrainfo",
		}

		respCreate, err := srv.CreateTrade(wctx, msgCreate)
		require.NoError(err)
		require.Equal(i, int(respCreate.Id))

		msgFulfill := &types.MsgFulfillTrade{
			Creator:         creatorB,
			Id:              respCreate.Id,
			CoinInputsIndex: 0,
			Items:           nil,
		}

		_, err = srv.FulfillTrade(wctx, msgFulfill)
		require.NoError(err)
	}
}

func (suite *IntegrationTestSuite) TestFulfillTradeMsgServerSimple2() {
	k := suite.k
	ctx := suite.ctx
	require := suite.Require()

	wctx := sdk.WrapSDKContext(ctx)
	srv := keeper.NewMsgServerImpl(k)

	privKey := ed25519.GenPrivKey()
	//creatorA := types.GenTestBech32FromString("creatorA")
	creatorB := types.GenTestBech32FromString("creatorB")

	msgCreate := &types.MsgCreateTrade{
		Creator:     creatorB,
		CoinInputs:  nil,
		ItemInputs:  nil,
		CoinOutputs: nil,
		ItemOutputs: nil,
		ExtraInfo:   "extrainfo",
	}
	for index, tc := range []struct {
		desc                         string
		msgFulfill                   types.MsgFulfillTrade
		updateIdMsgFulfill           bool
		msgCreate                    types.MsgCreateTrade
		updateCoinInputsMsgCreate    bool
		updatePaymentInfosForProcess bool
		valid                        bool
	}{
		{
			desc: "Trade does not exist",
			msgFulfill: types.MsgFulfillTrade{
				Creator:         creatorB,
				Id:              0,
				CoinInputsIndex: 0,
				Items:           nil,
			},
			updateIdMsgFulfill:           true,
			msgCreate:                    *msgCreate,
			updateCoinInputsMsgCreate:    false,
			updatePaymentInfosForProcess: false,
			valid:                        false,
		},
		{
			desc: "Invalid coinInputs index",
			msgFulfill: types.MsgFulfillTrade{
				Creator:         creatorB,
				Id:              0,
				CoinInputsIndex: 2,
				Items:           nil,
			},
			updateIdMsgFulfill:           false,
			msgCreate:                    *msgCreate,
			updateCoinInputsMsgCreate:    true,
			updatePaymentInfosForProcess: false,
			valid:                        false,
		},
		{
			desc: "Invalid PaymentInfo",
			msgFulfill: types.MsgFulfillTrade{
				Creator:         creatorB,
				Id:              0,
				CoinInputsIndex: 0,
				Items:           nil,
				PaymentInfos: []types.PaymentInfo{
					{
						PurchaseId:    "1",
						ProcessorName: "test",
						PayerAddr:     "pylon0123",
						Amount:        sdk.OneInt(),
						ProductId:     "1",
						Signature:     "test",
					},
				},
			},
			updateIdMsgFulfill:           false,
			msgCreate:                    *msgCreate,
			updateCoinInputsMsgCreate:    false,
			updatePaymentInfosForProcess: false,
			valid:                        false,
		},
		{
			desc: "Process PaymentInfo error",
			msgFulfill: types.MsgFulfillTrade{
				Creator:         creatorB,
				Id:              0,
				CoinInputsIndex: 0,
				Items:           nil,
				PaymentInfos: []types.PaymentInfo{{
					PurchaseId:    "test",
					ProcessorName: "TestPayment",
					PayerAddr:     types.GenTestBech32FromString(types.TestCreator),
					Amount:        sdk.NewInt(2),
					ProductId:     "testProductId",
					Signature:     genTestPaymentInfoSignature("testPurchaseId", types.GenTestBech32FromString(types.TestCreator), "testProductId", sdk.NewInt(2), privKey),
				}},
			},
			updateIdMsgFulfill:           false,
			msgCreate:                    *msgCreate,
			updateCoinInputsMsgCreate:    false,
			updatePaymentInfosForProcess: true,
			valid:                        false,
		},
		{
			desc: "Not enough balance to pay for trade coinInputs",
			msgFulfill: types.MsgFulfillTrade{
				Creator:         creatorB,
				Id:              0,
				CoinInputsIndex: 0,
				Items:           nil,
			},
			updateIdMsgFulfill:           false,
			msgCreate:                    *msgCreate,
			updateCoinInputsMsgCreate:    true,
			updatePaymentInfosForProcess: false,
			valid:                        false,
		},
		{
			desc: "Valid",
			msgFulfill: types.MsgFulfillTrade{
				Creator:         creatorB,
				Id:              0,
				CoinInputsIndex: 0,
				Items:           nil,
			},
			updateIdMsgFulfill:           false,
			msgCreate:                    *msgCreate,
			updateCoinInputsMsgCreate:    false,
			updatePaymentInfosForProcess: false,
			valid:                        true,
		},
	} {
		suite.Run(tc.desc, func() {
			if tc.updateIdMsgFulfill {
				tc.msgFulfill.Id = math.MaxUint64
			} else {
				tc.msgFulfill.Id = uint64(index)
			}
			if tc.updateCoinInputsMsgCreate {
				tc.msgCreate.CoinInputs = []types.CoinInput{{
					Coins: sdk.NewCoins(
						sdk.NewCoin(types.PylonsCoinDenom, sdk.NewInt(10)),
					),
				}}
			} else {
				tc.msgCreate.CoinInputs = nil
			}

			if tc.updatePaymentInfosForProcess {
				params := k.GetParams(suite.ctx)
				params.PaymentProcessors = append(params.PaymentProcessors, types.PaymentProcessor{
					CoinDenom:            types.PylonsCoinDenom,
					PubKey:               base64.StdEncoding.EncodeToString(privKey.PubKey().Bytes()),
					ProcessorPercentage:  types.DefaultProcessorPercentage,
					ValidatorsPercentage: types.DefaultValidatorsPercentage,
					Name:                 "TestPayment",
				})
				k.SetParams(suite.ctx, params)
			}
			respCreate, err := srv.CreateTrade(wctx, &tc.msgCreate)
			require.NoError(err)
			require.Equal(index, int(respCreate.Id))

			_, err = srv.FulfillTrade(wctx, &tc.msgFulfill)
			if tc.valid {
				require.NoError(err)
			} else {
				require.Error(err)
			}
		})
	}
}
