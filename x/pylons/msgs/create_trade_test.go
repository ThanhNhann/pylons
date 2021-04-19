package msgs

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/Pylons-tech/pylons/x/pylons/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestCreateTradeGetSignBytesItemInput(t *testing.T) {
	sdkAddr, err := sdk.AccAddressFromBech32("cosmos1y8vysg9hmvavkdxpvccv2ve3nssv5avm0kt337")
	require.NoError(t, err)
	msg := NewMsgCreateTrade(
		types.CoinInputList{},
		types.GenTradeItemInputList("UTestCreateTrade-CB-001", []string{"Raichu"}),
		types.NewPylon(10),
		types.ItemList{},
		"Test CreateTrade GetSignBytes",
		sdkAddr.String())
	err = msg.ValidateBasic()
	require.NoError(t, err)

	expectedSignBytes := `{
      "CoinInputs": [],
      "CoinOutputs": [
        {
          "amount": "10",
          "denom": "pylon"
        }
      ],
      "ExtraInfo": "Test CreateTrade GetSignBytes",
      "ItemInputs": [
        {
          "CookbookID": "UTestCreateTrade-CB-001",
          "ItemInput": {
            "Conditions":{"Doubles":{"params":null},"Longs":{"List":null},"Strings":{"List":null}},
            "Doubles": {
              "params": null
            },
            "ID": "Raichu",
            "Longs": {
              "List": null
            },
            "Strings": {
              "List": [
                {
                  "Key": "Name",
                  "Value": "Raichu"
                }
              ]
            },
            "TransferFee": {
              "MaxValue": 10000
            }
          }
        }
      ],
      "ItemOutputs": [],
      "Sender": "cosmos1y8vysg9hmvavkdxpvccv2ve3nssv5avm0kt337"
    }`
	buffer := new(bytes.Buffer)
	err = json.Compact(buffer, []byte(expectedSignBytes))
	require.NoError(t, err)
	require.Equal(t, string(msg.GetSignBytes()), buffer.String())
}

func TestCreateTradeGetSignBytesUnorderedCoinInputs(t *testing.T) {
	sdkAddr, err := sdk.AccAddressFromBech32("cosmos1y8vysg9hmvavkdxpvccv2ve3nssv5avm0kt337")
	require.NoError(t, err)
	msg := NewMsgCreateTrade(
		types.CoinInputList{
			{Coin: "aaaa", Count: 100},
			{Coin: "zzzz", Count: 100},
			{Coin: "cccc", Count: 100},
		},
		types.TradeItemInputList{},
		types.NewPylon(10),
		types.ItemList{},
		"Test CreateTrade GetSignBytes",
		sdkAddr.String())
	err = msg.ValidateBasic()
	require.NoError(t, err)

	expectedSignBytes := `{
      "CoinInputs": [
        {
          "Coin": "aaaa",
          "Count": 100
        },
        {
          "Coin": "zzzz",
          "Count": 100
        },
        {
          "Coin": "cccc",
          "Count": 100
        }
      ],
      "CoinOutputs": [
        {
          "amount": "10",
          "denom": "pylon"
        }
      ],
      "ExtraInfo": "Test CreateTrade GetSignBytes",
      "ItemInputs": [],
      "ItemOutputs": [],
      "Sender": "cosmos1y8vysg9hmvavkdxpvccv2ve3nssv5avm0kt337"
    }`
	buffer := new(bytes.Buffer)
	err = json.Compact(buffer, []byte(expectedSignBytes))
	require.NoError(t, err)
	require.Equal(t, string(msg.GetSignBytes()), buffer.String(), string(msg.GetSignBytes()))
}
