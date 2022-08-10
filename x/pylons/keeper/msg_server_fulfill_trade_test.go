package keeper_test

import (
	"encoding/base64"
	"fmt"
	"math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Pylons-tech/pylons/x/pylons/keeper"
	"github.com/Pylons-tech/pylons/x/pylons/types"
)

func (suite *IntegrationTestSuite) TestFulfillTradeMsgServerSimple2() {
	k := suite.k
	ctx := suite.ctx
	require := suite.Require()
	bk := suite.bankKeeper

	wctx := sdk.WrapSDKContext(ctx)
	srv := keeper.NewMsgServerImpl(k)

	privKey := ed25519.GenPrivKey()
	creatorA := types.GenTestBech32FromString("creator1")
	creatorB := types.GenTestBech32FromString("creator2")

	trashStr := types.GenTestBech32FromString("trash")
	trashAddress := sdk.MustAccAddressFromBech32(trashStr)

	msgCreate := &types.MsgCreateTrade{
		Creator:     creatorA,
		CoinInputs:  nil,
		ItemInputs:  nil,
		CoinOutputs: nil,
		ItemOutputs: nil,
		ExtraInfo:   "extrainfo",
	}
	for index, tc := range []struct {
		desc                            string
		msgFulfill                      types.MsgFulfillTrade
		updateIdMsgFulfill              bool
		msgCreate                       types.MsgCreateTrade
		updateCoinInputsMsgCreate       bool
		updatePaymentInfosForProcess    bool
		updateItemOutputsMsgCreate      bool
		setItem                         bool
		tradeableOfItem                 bool
		setUpPayFeeForTradeFulFill      bool
		setUpSendCoinsToCookBookCreator bool
		setUpPayFeeForTradeCreate       bool
		setTradePercentage              bool
		valid                           bool
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
			updateItemOutputsMsgCreate:   false,
			setItem:                      false,
			tradeableOfItem:              false,
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
			updateItemOutputsMsgCreate:   false,
			setItem:                      false,
			tradeableOfItem:              false,
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
			updateItemOutputsMsgCreate:   false,
			setItem:                      false,
			tradeableOfItem:              false,
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
			updateItemOutputsMsgCreate:   false,
			setItem:                      false,
			tradeableOfItem:              false,
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
			updateItemOutputsMsgCreate:   false,
			setItem:                      false,
			tradeableOfItem:              false,
			valid:                        false,
		},
		{
			desc: "Error match ItemInputs for trade",
			msgFulfill: types.MsgFulfillTrade{
				Creator:         creatorB,
				Id:              0,
				CoinInputsIndex: 0,
				Items: []types.ItemRef{
					{
						CookbookId: "test",
						ItemId:     "test",
					},
				},
			},
			updateIdMsgFulfill:           false,
			msgCreate:                    *msgCreate,
			updateCoinInputsMsgCreate:    false,
			updatePaymentInfosForProcess: false,
			updateItemOutputsMsgCreate:   false,
			setItem:                      false,
			tradeableOfItem:              false,
			valid:                        false,
		},
		{
			desc: "Cannot be traded with item and cookbook id",
			msgFulfill: types.MsgFulfillTrade{
				Creator:         creatorB,
				Id:              0,
				CoinInputsIndex: 0,
				Items: []types.ItemRef{
					{
						CookbookId: "",
						ItemId:     "",
					},
				},
			},
			updateIdMsgFulfill:           false,
			msgCreate:                    *msgCreate,
			updateCoinInputsMsgCreate:    false,
			updatePaymentInfosForProcess: false,
			updateItemOutputsMsgCreate:   false,
			setItem:                      true,
			tradeableOfItem:              false,
			valid:                        false,
		},
		{
			desc: "Cannot use CoinOutputs to pay for the items provided",
			msgFulfill: types.MsgFulfillTrade{
				Creator:         creatorB,
				Id:              0,
				CoinInputsIndex: 0,
				Items: []types.ItemRef{
					{
						CookbookId: "",
						ItemId:     "",
					},
				},
			},
			updateIdMsgFulfill:           false,
			msgCreate:                    *msgCreate,
			updateCoinInputsMsgCreate:    false,
			updatePaymentInfosForProcess: false,
			updateItemOutputsMsgCreate:   false,
			setItem:                      true,
			tradeableOfItem:              true,
			valid:                        false,
		},
		{
			desc: "CoinInputs not sufficient to pay transfer fees",
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
			updateItemOutputsMsgCreate:   true,
			setItem:                      false,
			tradeableOfItem:              false,
			valid:                        false,
		},
		{
			desc: "Insufficient fees from Trade FulFill to pay for inputChain",
			msgFulfill: types.MsgFulfillTrade{
				Creator:         creatorB,
				Id:              0,
				CoinInputsIndex: 0,
				Items: []types.ItemRef{
					{
						CookbookId: "",
						ItemId:     "",
					},
				},
			},
			updateIdMsgFulfill:           false,
			msgCreate:                    *msgCreate,
			updateCoinInputsMsgCreate:    false,
			updatePaymentInfosForProcess: false,
			updateItemOutputsMsgCreate:   false,
			setItem:                      false,
			tradeableOfItem:              false,
			setUpPayFeeForTradeFulFill:   true,
			setTradePercentage:           true,
			valid:                        false,
		},
		{
			desc: "Insufficient fees from Trade FulFill to send coins to cookbook's creator",
			msgFulfill: types.MsgFulfillTrade{
				Creator:         creatorB,
				Id:              0,
				CoinInputsIndex: 0,
				Items: []types.ItemRef{
					{},
					{},
				},
			},
			updateIdMsgFulfill:              false,
			msgCreate:                       *msgCreate,
			updateCoinInputsMsgCreate:       false,
			updatePaymentInfosForProcess:    false,
			updateItemOutputsMsgCreate:      false,
			tradeableOfItem:                 false,
			setUpSendCoinsToCookBookCreator: true,
			setTradePercentage:              true,
			valid:                           false,
		},
		{
			desc: "Insufficient fees from Trade Creator to pay for outputChain",
			msgFulfill: types.MsgFulfillTrade{
				Creator:         creatorB,
				Id:              0,
				CoinInputsIndex: 0,
				Items: []types.ItemRef{
					{},
				},
			},
			updateIdMsgFulfill:           false,
			msgCreate:                    *msgCreate,
			updateCoinInputsMsgCreate:    false,
			updatePaymentInfosForProcess: false,
			updateItemOutputsMsgCreate:   false,
			setItem:                      false,
			tradeableOfItem:              false,
			setUpPayFeeForTradeCreate:    true,
			setTradePercentage:           true,
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
			updateItemOutputsMsgCreate:   false,
			setItem:                      false,
			tradeableOfItem:              false,
			valid:                        true,
		},
	} {
		suite.Run(tc.desc, func() {

			tradeCreatorAddr, _ := sdk.AccAddressFromBech32(tc.msgCreate.Creator)
			k.SetPylonsAccount(ctx, types.AccountAddr{Value: tradeCreatorAddr.String()}, types.Username{Value: tc.msgCreate.Creator})

			tradeFulfillerAddr, _ := sdk.AccAddressFromBech32(tc.msgFulfill.Creator)
			k.SetPylonsAccount(ctx, types.AccountAddr{Value: tradeFulfillerAddr.String()}, types.Username{Value: tc.msgFulfill.Creator})
			if tc.msgCreate.ItemOutputs != nil {
				tc.msgCreate.ItemOutputs = nil
			}
			//Begin config
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

			if tc.updateItemOutputsMsgCreate {
				tc.msgCreate.ItemOutputs = []types.ItemRef{
					{
						CookbookId: fmt.Sprintf("%d", index),
						ItemId:     fmt.Sprintf("%d", index),
					},
				}
				item := &types.Item{
					Owner:      creatorA,
					CookbookId: fmt.Sprintf("%d", index),
					Id:         fmt.Sprintf("%d", index),
					Tradeable:  true,
				}

				k.SetItem(ctx, *item)
			}

			if tc.setItem {

				tc.msgFulfill.Items[0].CookbookId = fmt.Sprintf("%d", index)
				tc.msgFulfill.Items[0].ItemId = fmt.Sprintf("%d", index)

				tc.msgCreate.ItemInputs = []types.ItemInput{
					{
						Id: fmt.Sprintf("%d", index),
					},
				}
				item := &types.Item{
					Owner:      creatorB,
					CookbookId: fmt.Sprintf("%d", index),
					Id:         fmt.Sprintf("%d", index),
				}
				if tc.tradeableOfItem {
					item.Tradeable = true
				} else {
					item.Tradeable = false
				}
				k.SetItem(ctx, *item)
			}

			if tc.setUpPayFeeForTradeFulFill {
				err := k.MintCoinsToAddr(ctx, tradeCreatorAddr, sdk.NewCoins(sdk.NewCoin("coin0", sdk.NewInt(4))))
				require.NoError(err)
				err = k.MintCoinsToAddr(ctx, tradeCreatorAddr, sdk.NewCoins(sdk.NewCoin("coin1", sdk.NewInt(10))))
				require.NoError(err)

				tc.msgFulfill.Items[0].CookbookId = fmt.Sprintf("%d", index)
				tc.msgFulfill.Items[0].ItemId = fmt.Sprintf("%d", index)
				tc.msgCreate.ItemInputs = []types.ItemInput{
					{
						Id: fmt.Sprintf("%d", index),
					},
				}
				item := &types.Item{
					Owner:       creatorB,
					CookbookId:  fmt.Sprintf("%d", index),
					Id:          fmt.Sprintf("%d", index),
					TransferFee: sdk.Coins{sdk.NewCoin("coin0", sdk.NewInt(10)), sdk.NewCoin("coin1", sdk.NewInt(10))},
					Tradeable:   true,
				}
				tc.msgCreate.CoinOutputs = sdk.Coins{sdk.NewCoin("coin0", sdk.NewInt(4)), sdk.NewCoin("coin1", sdk.NewInt(10))}
				if tc.setTradePercentage {
					item.TradePercentage = sdk.OneDec()
				}
				k.SetItem(ctx, *item)
			}

			if tc.setUpSendCoinsToCookBookCreator {
				err := k.MintCoinsToAddr(ctx, tradeCreatorAddr, sdk.NewCoins(sdk.NewCoin("coin0", sdk.NewInt(10))))
				require.NoError(err)
				err = k.MintCoinsToAddr(ctx, tradeCreatorAddr, sdk.NewCoins(sdk.NewCoin("coin2", sdk.NewInt(4))))
				require.NoError(err)
				err = k.MintCoinsToAddr(ctx, tradeCreatorAddr, sdk.NewCoins(sdk.NewCoin("coin4", sdk.NewInt(5))))
				require.NoError(err)

				err = k.MintCoinsToAddr(ctx, tradeFulfillerAddr, sdk.NewCoins(sdk.NewCoin("coin0", sdk.NewInt(1))))
				require.NoError(err)

				tc.msgFulfill.Items[0].CookbookId = fmt.Sprintf("%d", index)
				tc.msgFulfill.Items[0].ItemId = fmt.Sprintf("%d", index)

				tc.msgFulfill.Items[1].CookbookId = fmt.Sprintf("%d", index) + "test"
				tc.msgFulfill.Items[1].ItemId = fmt.Sprintf("%d", index) + "test"

				cookbook := &types.MsgCreateCookbook{
					Creator:      creatorA,
					Id:           fmt.Sprintf("%d", index),
					Name:         fmt.Sprintf("%d", index),
					Description:  "descdescdescdescdescdesc",
					Developer:    "",
					Version:      "v0.0.1",
					SupportEmail: "test@email.com",
					Enabled:      true,
				}
				_, err = srv.CreateCookbook(ctx, cookbook)
				require.NoError(err)

				cookbook2 := &types.MsgCreateCookbook{
					Creator:      creatorA,
					Id:           fmt.Sprintf("%d", index) + "test",
					Name:         fmt.Sprintf("%d", index) + "test",
					Description:  "descdescdescdescdescdesc",
					Developer:    "",
					Version:      "v0.0.1",
					SupportEmail: "test@email.com",
					Enabled:      true,
				}
				_, err = srv.CreateCookbook(ctx, cookbook2)
				require.NoError(err)
				tc.msgCreate.ItemInputs = []types.ItemInput{
					{
						Id: fmt.Sprintf("%d", index),
					},
					{
						Id: fmt.Sprintf("%d", index) + "test",
					},
				}

				items := []types.Item{
					{
						Owner:           creatorB,
						CookbookId:      fmt.Sprintf("%d", index),
						Id:              fmt.Sprintf("%d", index),
						TransferFee:     sdk.Coins{sdk.NewCoin("coin0", sdk.NewInt(10)), sdk.NewCoin("coin1", sdk.NewInt(10)), sdk.NewCoin("coin2", sdk.NewInt(10))},
						TradePercentage: sdk.OneDec(),
						Tradeable:       true,
					},
					{
						Owner:           creatorB,
						CookbookId:      fmt.Sprintf("%d", index) + "test",
						Id:              fmt.Sprintf("%d", index) + "test",
						TransferFee:     sdk.Coins{sdk.NewCoin("coin2", sdk.NewInt(4)), sdk.NewCoin("coin3", sdk.NewInt(5)), sdk.NewCoin("coin4", sdk.NewInt(6))},
						TradePercentage: sdk.OneDec(),
						Tradeable:       true,
					},
				}
				tc.msgCreate.CoinOutputs = sdk.Coins{sdk.NewCoin("coin0", sdk.NewInt(10)), sdk.NewCoin("coin2", sdk.NewInt(4)), sdk.NewCoin("coin4", sdk.NewInt(5))}

				for _, item := range items {
					k.SetItem(ctx, item)
				}
			}

			if tc.setUpPayFeeForTradeCreate {
				err := k.MintCoinsToAddr(ctx, tradeCreatorAddr, sdk.NewCoins(sdk.NewCoin("coin0", sdk.NewInt(4))))
				require.NoError(err)
				err = k.MintCoinsToAddr(ctx, tradeCreatorAddr, sdk.NewCoins(sdk.NewCoin("coin1", sdk.NewInt(10))))
				require.NoError(err)

				err = k.MintCoinsToAddr(ctx, tradeFulfillerAddr, sdk.NewCoins(sdk.NewCoin("coin1", sdk.NewInt(10))))
				require.NoError(err)

				cookbook := &types.MsgCreateCookbook{
					Creator:      creatorA,
					Id:           fmt.Sprintf("%d", index),
					Name:         fmt.Sprintf("%d", index),
					Description:  "descdescdescdescdescdesc",
					Developer:    "",
					Version:      "v0.0.1",
					SupportEmail: "test@email.com",
					Enabled:      true,
				}
				_, err = srv.CreateCookbook(ctx, cookbook)
				require.NoError(err)

				tc.msgFulfill.Items[0].CookbookId = fmt.Sprintf("%d", index)
				tc.msgFulfill.Items[0].ItemId = fmt.Sprintf("%d", index)
				tc.msgCreate.ItemInputs = []types.ItemInput{
					{
						Id: fmt.Sprintf("%d", index),
					},
				}
				item := &types.Item{
					Owner:       creatorB,
					CookbookId:  fmt.Sprintf("%d", index),
					Id:          fmt.Sprintf("%d", index),
					TransferFee: sdk.Coins{sdk.NewCoin("coin0", sdk.NewInt(10)), sdk.NewCoin("coin1", sdk.NewInt(10))},
					Tradeable:   true,
				}
				tc.msgCreate.ItemOutputs = []types.ItemRef{
					{
						CookbookId: fmt.Sprintf("%d", index),
						ItemId:     fmt.Sprintf("%d", index),
					},
				}

				tc.msgCreate.CoinOutputs = sdk.Coins{sdk.NewCoin("coin0", sdk.NewInt(4)), sdk.NewCoin("coin1", sdk.NewInt(10))}
				if tc.setTradePercentage {
					item.TradePercentage = sdk.OneDec()
				}
				k.SetItem(ctx, *item)
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
			//End config

			respCreate, err := srv.CreateTrade(wctx, &tc.msgCreate)
			require.NoError(err)
			require.Equal(index, int(respCreate.Id))

			_, err = srv.FulfillTrade(wctx, &tc.msgFulfill)
			if tc.valid {
				require.NoError(err)
			} else {
				require.Error(err)
			}
			if bk.SpendableCoins(ctx, tradeCreatorAddr) != nil {
				suite.pylonsApp.BankKeeper.SendCoins(ctx, tradeCreatorAddr, trashAddress, bk.SpendableCoins(ctx, tradeCreatorAddr))
			}
			if bk.SpendableCoins(ctx, tradeFulfillerAddr) != nil {
				suite.pylonsApp.BankKeeper.SendCoins(ctx, tradeFulfillerAddr, trashAddress, bk.SpendableCoins(ctx, tradeFulfillerAddr))
			}
		})
	}
}
