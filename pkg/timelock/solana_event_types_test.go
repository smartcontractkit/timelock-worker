package timelock

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/mr-tron/base58"
	"github.com/stretchr/testify/require"
)

func makeTxWithMeta(t *testing.T, fullJSON string) *rpc.TransactionWithMeta {
	t.Helper()

	var parsed struct {
		Result struct {
			Slot        uint64               `json:"slot"`
			Meta        *rpc.TransactionMeta `json:"meta"`
			Transaction struct {
				Signatures []string       `json:"signatures"`
				Message    solana.Message `json:"message"`
			} `json:"transaction"`
		} `json:"result"`
	}
	require.NoError(t, json.Unmarshal([]byte(fullJSON), &parsed))

	// Rebuild transaction
	tx := &solana.Transaction{
		Message:    parsed.Result.Transaction.Message,
		Signatures: make([]solana.Signature, len(parsed.Result.Transaction.Signatures)),
	}
	for i, sigStr := range parsed.Result.Transaction.Signatures {
		sig, err := solana.SignatureFromBase58(sigStr)
		require.NoError(t, err)
		tx.Signatures[i] = sig
	}

	binTx, err := tx.MarshalBinary()
	require.NoError(t, err)

	// Encode to base58
	b58 := base58.Encode(binTx)

	// Marshal as JSON string (e.g., `"base58value"`)
	jsonBz, err := json.Marshal(b58)
	require.NoError(t, err)

	// Unmarshal into DataBytesOrJSON
	var txField rpc.DataBytesOrJSON
	require.NoError(t, json.Unmarshal(jsonBz, &txField))

	return &rpc.TransactionWithMeta{
		Slot:        parsed.Result.Slot,
		Meta:        parsed.Result.Meta,
		Transaction: &txField,
	}
}

func TestParseTimelockEvents_CallExecuted(t *testing.T) {
	const executeBatchTxJSON = `{
  "jsonrpc": "2.0",
  "result": {
    "blockTime": 1748021125,
    "meta": {
      "computeUnitsConsumed": 27191,
      "err": null,
      "fee": 5000,
      "innerInstructions": [
        {
          "index": 0,
          "instructions": [
            {
              "accounts": [
                2,
                5
              ],
              "data": "VnVctNXAKG1",
              "programIdIndex": 7,
              "stackHeight": 2
            }
          ]
        }
      ],
      "loadedAddresses": {
        "readonly": [],
        "writable": []
      },
      "logMessages": [
        "Program DoajfR5tK24xVw51fWcawUZWhAXD8yrBJVacc13neVQA invoke [1]",
        "Program log: Instruction: ExecuteBatch",
        "Program offqSMQWgQud6WJz694LRzkeN5kMYpCHTpXQr3Rkcjm invoke [2]",
        "Program log: Instruction: AcceptOwnership",
        "Program data: rD3Nt/oyJmJlE+rTgDo96ayxRLzURrZ0VXRhHG1wd0AEDOX6PDYh3R74fnS+SLFE6OvgAxbo+S+KqA2nm/gNrv6H0f98BW1Y",
        "Program offqSMQWgQud6WJz694LRzkeN5kMYpCHTpXQr3Rkcjm consumed 3872 of 1379263 compute units",
        "Program offqSMQWgQud6WJz694LRzkeN5kMYpCHTpXQr3Rkcjm success",
        "Program data: 7Xjujr0lQYCbMz3bJrZjB2awZbsPIBEDQgI0hHOTP3L79epG611RUgAAAAAAAAAAC/Rw6A3Ml4d6nU2Njm4F9OFrhELkYDVH0V5R8rVYXtQIAAAArBcrDe7VVZY=",
        "Program DoajfR5tK24xVw51fWcawUZWhAXD8yrBJVacc13neVQA consumed 27041 of 1400000 compute units",
        "Program DoajfR5tK24xVw51fWcawUZWhAXD8yrBJVacc13neVQA success",
        "Program ComputeBudget111111111111111111111111111111 invoke [1]",
        "Program ComputeBudget111111111111111111111111111111 success"
      ],
      "postBalances": [
        22587136898,
        2540400,
        13752960,
        2540400,
        9744000,
        10000000000,
        15701760,
        1141440,
        1141440,
        1
      ],
      "postTokenBalances": [],
      "preBalances": [
        22587141898,
        2540400,
        13752960,
        2540400,
        9744000,
        10000000000,
        15701760,
        1141440,
        1141440,
        1
      ],
      "preTokenBalances": [],
      "rewards": [],
      "status": {
        "Ok": null
      }
    },
    "slot": 382828012,
    "transaction": [
      "KCUrq6ztDiyKw9y7sWgbPKouCvFx6Zd1XGGcNwWUFNWPPHNu1wt29JEa96qzedvqnaEgKj3LRtnhHH8Y8Nq6MzDUmwhVPmSkpATPpuZZua1RfAh6Gba3ReBcQ5rKrMJFsxMJ1kmx87For3uWsT6WFnViAUCM4oLagQvZubzAQnVxcec7JVJjowAp2sGe6QLJzKpxCZZ29EV8SDTRHh7AwnZVMmLsBFNKNnfmtSDskqbvqCHa2azFsQYqHmhGj5VdaJZUKZiTXzbn1LnB97KPsbMzjRBumuyVAKgx3EUQkpCVkDgiqm3pcHpdeU4Lp9M3R9F6rh9fkhLTuDpX8vsSEtvg1gyp4oVkFFKvpd7m1KLVbvPCg54XY4eLNjb7AD1Him6dosLBUqFz35yDjHqcpmMaaYBS28ycUqBg8mCM3zaXPRTM2nfr8dv9k9EfV2yqLSTybYnYroSw6mLwpDUQB4bmDrqXXMAhCwD6niAYPgpcxRTMx7k3eZAVXXtG6QBzEATKER8TWnJuGkMBzEt8fEnmzx9PDQX6T19QSueiGoaJvoi64ubifeyFj19Lh6wD9jk1CKmjnhGjwgVyGv7FcUPrMpuHWYeCeaZwMgb3BEwJ2XYPfejKwXeohaReHg1agRePsWNSpGFhf9gs4xhRrGgeSegaY47wJ7dbub9LBWnFdCSbUS3zPFv1EQucT",
      "base58"
    ],
    "version": "legacy"
  },
  "id": 1
}`

	var rawResponse struct {
		Result rpc.TransactionWithMeta `json:"result"`
	}
	err := json.Unmarshal([]byte(executeBatchTxJSON), &rawResponse)

	// --- act ---
	events, err := ParseTimelockEvents(&rawResponse.Result)
	require.NoError(t, err)

	// --- assert ---
	require.Len(t, events.Executed, 1)
	require.Len(t, events.Scheduled, 0)
	require.Len(t, events.Cancelled, 0)

	event := events.Executed[0]
	require.Equal(t, "offqSMQWgQud6WJz694LRzkeN5kMYpCHTpXQr3Rkcjm", event.Target.String())
	require.Equal(t, "\xac\x17+\r\xee\xd5U\x96", string(event.Data))
	require.Equal(t, uint64(0), event.Index)
	require.Equal(t, operationKeyFromHex("0x9b333ddb26b6630766b065bb0f2011034202348473933f72fbf5ea46eb5d5152"), event.ID)
}

func TestParseTimelockEvents_CallScheduledTx(t *testing.T) {
	const scheduleBatchTxJSON = `{
  "jsonrpc": "2.0",
  "result": {
    "blockTime": 1748436483,
    "meta": {
      "computeUnitsConsumed": 64703,
      "err": null,
      "fee": 5000,
      "innerInstructions": [
        {
          "index": 0,
          "instructions": [
            {
              "accounts": [
                3,
                7,
                8,
                4
              ],
              "data": "3eJCB4YtuS2me8jfjvQLeGjPBszgEL1WoYhvifHfLVTcKd85pZKEqjhZfZPF838bCTG6weKvDikU3kxpZQ516gyPVdGZJLNWs3LXT5Q31R9iLo",
              "programIdIndex": 6,
              "stackHeight": 2
            }
          ]
        }
      ],
      "loadedAddresses": {
        "readonly": [],
        "writable": []
      },
      "logMessages": [
        "Program 5vNJx78mz7KVMjhuipyr9jKBKcMrKYGdjGkgE4LUmjKk invoke [1]",
        "Program log: Instruction: Execute",
        "Program DoajfR5tK24xVw51fWcawUZWhAXD8yrBJVacc13neVQA invoke [2]",
        "Program log: Instruction: ScheduleBatch",
        "Program data: v1Vap4TfuDnIXpa3xd/TcOPVES/wWdVelcwplTNt8T83mTbUEzWKeAAAAAAAAAAArJcpd+G6tf/2xoYpH5lP9S80q+0bx9+I929D+SZZLI8AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAGg5/GkAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAALAEAAAAAAAAoAAAA2iWLa47kM9se+H50vkixROjr4AMW6PkviqgNp5v4Da7+h9H/fAVtWA==",
        "Program data: v1Vap4TfuDnIXpa3xd/TcOPVES/wWdVelcwplTNt8T83mTbUEzWKeAEAAAAAAAAArJcpd+G6tf/2xoYpH5lP9S80q+0bx9+I929D+SZZLI8AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAGg5/GkAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAALAEAAAAAAAAoAAAA2iWLa47kM9se+H50vkixROjr4AMW6PkviqgNp5v4Da7+h9H/fAVtWA==",
        "Program data: v1Vap4TfuDnIXpa3xd/TcOPVES/wWdVelcwplTNt8T83mTbUEzWKeAIAAAAAAAAArJcpd+G6tf/2xoYpH5lP9S80q+0bx9+I929D+SZZLI8AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAGg5/GkAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAALAEAAAAAAAAIAAAAavAQrYnVo/Y=",
        "Program data: v1Vap4TfuDnIXpa3xd/TcOPVES/wWdVelcwplTNt8T83mTbUEzWKeAMAAAAAAAAArJcpd+G6tf/2xoYpH5lP9S80q+0bx9+I929D+SZZLI8AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAGg5/GkAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAALAEAAAAAAAAIAAAAavAQrYnVo/Y=",
        "Program data: v1Vap4TfuDnIXpa3xd/TcOPVES/wWdVelcwplTNt8T83mTbUEzWKeAQAAAAAAAAArJcpd+G6tf/2xoYpH5lP9S80q+0bx9+I929D+SZZLI8AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAGg5/GkAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAALAEAAAAAAAAPAAAAdx4OtHPhp+4DAAAAAwQH",
        "Program data: v1Vap4TfuDnIXpa3xd/TcOPVES/wWdVelcwplTNt8T83mTbUEzWKeAUAAAAAAAAArJcpd+G6tf/2xoYpH5lP9S80q+0bx9+I929D+SZZLI8AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAGg5/GkAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAALAEAAAAAAAAPAAAAdx4OtHPhp+4DAAAAAwQH",
        "Program DoajfR5tK24xVw51fWcawUZWhAXD8yrBJVacc13neVQA consumed 25428 of 1367858 compute units",
        "Program DoajfR5tK24xVw51fWcawUZWhAXD8yrBJVacc13neVQA success",
        "Program data: 3Q/UHSP8/05UAAAAAAAAAL47GvGrhOB/P+m1o/57yJMpAL8r+oYfe0/EYkBx0NiHUAAAAPKMV2pH4lYgRTRSNk53ZzFLOFp2aTZNY0xka2FHREQ1Q2xYMUtreVbIXpa3xd/TcOPVES/wWdVelcwplTNt8T83mTbUEzWKeCwBAAAAAAAA",
        "Program 5vNJx78mz7KVMjhuipyr9jKBKcMrKYGdjGkgE4LUmjKk consumed 64553 of 1400000 compute units",
        "Program 5vNJx78mz7KVMjhuipyr9jKBKcMrKYGdjGkgE4LUmjKk success",
        "Program ComputeBudget111111111111111111111111111111 invoke [1]",
        "Program ComputeBudget111111111111111111111111111111 success"
      ],
      "postBalances": [
        14964655256,
        6737280,
        1252800,
        11379600,
        9951781120,
        1343280,
        1141440,
        9744000,
        15701760,
        1141440,
        1
      ],
      "postTokenBalances": [],
      "preBalances": [
        14964660256,
        6737280,
        1252800,
        11379600,
        9951781120,
        1343280,
        1141440,
        9744000,
        15701760,
        1141440,
        1
      ],
      "preTokenBalances": [],
      "rewards": [],
      "status": {
        "Ok": null
      }
    },
    "slot": 383874824,
    "transaction": [
      "3G7zdvGQnbcTxFckHHCSbs9ew2z4xs3VVveeXt3CqWPxHcB7K4P9pXJRjisaUKsrToku6DrBcW3m4R7N3NwHFwk43cwRuqf7pwgECLh5ym6LK1cqafbrve28wo7GC6d3V2C2n4iEFGgiPZKHn6A4qmWVsxrpzUm9Tc1yKJWFfgTeRevQLwe7xRQtXVifMkrQmMVGFnezsuKhy3GsnpejAP6fMtLSyMqu2EZ9gEVnz7qUTXFLEt3o5Ax2pDx1W3R5dYfDum9iAikjaDW6LZ9irQ6NqCAHcRjTCNobyfNdD6Mv9NiqUxGEGi2kCfiAaUwcnKqcSQwiuamrg7KUT7YkayMDTJ7GxQazyYY8jon4WxPsmdikJKA69AXJs2AvsFfp25Ji2qRDiqYXpxYH1uNj8cG79KxYc2L7HDdTi1g8qrkL8MURx5AeijekKs2vaHJfyiFwTZ5Fic7r8XudFEunXjDDabGhEFJViTFh7DT9eHd3ZuB81ZhWZXx3uNm39yKWjb7g3cW5NDrnhYD4ubkZAgfY5iU2tyoaxG9ktrqvXtx9qKZvuLjUUi8GWxTr4TM6mQcs7XD2wsgFwExWPpBm6gB1m23X1Zm9cPJAbSrhim13rNqgCfs16v3BSFMbJ3X5DKi2sikCwFPQ3uP655z1XqFErd8U1Ea64b739C3Ky2ieapnDapXxxxbjmrjyxcjGce6KmdxSB9DLrv71SNpRX3vwDqaNQSpf7EEsWNU95XBUsKMXMyRRWXA9QoHDrnKb4M8jHcqmv4TzJjeb32hLXpMo8EJLzKBVXVdBD8ZFHMqh8rWFPiX1f5LqB675YcNH28VcY9AmrNwvo5Z5CqYydX3kMBLCXJrrA259AxvV8rxDj4PfDFeR8qT5fGnbR1Q5DKsmvFhsH4yVX2v97YckFRnWqTZ2xi3jFUsGihdoAR52BBUry7pgUi5E33zxToXhpG5opQZPv4y5X1rdbFrXqxZBkZac8fhghD2TetsZGES7cuHBA9ZUmoBPVTe8f",
      "base58"
    ],
    "version": "legacy"
  },
  "id": 1
}`

	var rawResponse struct {
		Result rpc.TransactionWithMeta `json:"result"`
	}
	err := json.Unmarshal([]byte(scheduleBatchTxJSON), &rawResponse)
	require.NoError(t, err, "should unmarshal JSON into rpc.TransactionWithMeta")

	// --- act ---
	events, err := ParseTimelockEvents(&rawResponse.Result)
	require.NoError(t, err)

	// --- assert ---
	require.Len(t, events.Executed, 0)
	require.Len(t, events.Scheduled, 6)
	require.Len(t, events.Cancelled, 0)

	event := events.Scheduled[0]
	require.Equal(t, "Ccip842gzYHhvdDkSyi2YVCoAWPbYJoApMFzSxQroE9C", event.Target.String())
	require.Equal(t, "\xda%\x8bk\x8e\xe43\xdb\x1e\xf8~t\xbeH\xb1D\xe8\xeb\xe0\x03\x16\xe8\xf9/\x8a\xa8\r\xa7\x9b\xf8\r\xae\xfe\x87\xd1\xff|\x05mX", string(event.Data))
	require.Equal(t, uint64(0), event.Index)
	require.Equal(t, operationKeyFromHex("0xc85e96b7c5dfd370e3d5112ff059d55e95cc2995336df13f379936d413358a78"), event.ID)
	require.Equal(t, big.NewInt(383874824), event.BlockNumber)
	require.Equal(t, "2G3nDopUPP1b4B8woZWVNME2Qo3i9eyAB7n1B2QLHHgM8gPb1dTK3iLkCqFyktgK9Jm4mU3Ny9aBahrcHUSzyTjg", event.TxHash)
}

func TestParseTimelockEvents_CancelledEvent(t *testing.T) {
	const cancelTxJSON = `{
		"jsonrpc": "2.0",
		"result": {
			"blockTime": 1748500000,
			"meta": {
				"computeUnitsConsumed": 12345,
				"err": null,
				"fee": 5000,
				"innerInstructions": [],
				"loadedAddresses": {
					"readonly": [],
					"writable": []
				},
				"logMessages": [
					"Program XXXXXX invoke [1]",
					"Program log: Instruction: Cancel",
					"Program data: iBcqQY/p6i7ercD/7t6twP/u3q3A/+7ercD/7t6twP/u3q3A/+4SNFZ4",
					"Program XXXXXX success"
				],
				"postBalances": [ 1000000000, 200000000 ],
				"preBalances": [ 1000005000, 200000000 ],
				"postTokenBalances": [],
				"preTokenBalances": [],
				"rewards": [],
				"status": { "Ok": null }
			},
			"slot": 383999999,
			"transaction": [
			  "KCUrq6ztDiyKw9y7sWgbPKouCvFx6Zd1XGGcNwWUFNWPPHNu1wt29JEa96qzedvqnaEgKj3LRtnhHH8Y8Nq6MzDUmwhVPmSkpATPpuZZua1RfAh6Gba3ReBcQ5rKrMJFsxMJ1kmx87For3uWsT6WFnViAUCM4oLagQvZubzAQnVxcec7JVJjowAp2sGe6QLJzKpxCZZ29EV8SDTRHh7AwnZVMmLsBFNKNnfmtSDskqbvqCHa2azFsQYqHmhGj5VdaJZUKZiTXzbn1LnB97KPsbMzjRBumuyVAKgx3EUQkpCVkDgiqm3pcHpdeU4Lp9M3R9F6rh9fkhLTuDpX8vsSEtvg1gyp4oVkFFKvpd7m1KLVbvPCg54XY4eLNjb7AD1Him6dosLBUqFz35yDjHqcpmMaaYBS28ycUqBg8mCM3zaXPRTM2nfr8dv9k9EfV2yqLSTybYnYroSw6mLwpDUQB4bmDrqXXMAhCwD6niAYPgpcxRTMx7k3eZAVXXtG6QBzEATKER8TWnJuGkMBzEt8fEnmzx9PDQX6T19QSueiGoaJvoi64ubifeyFj19Lh6wD9jk1CKmjnhGjwgVyGv7FcUPrMpuHWYeCeaZwMgb3BEwJ2XYPfejKwXeohaReHg1agRePsWNSpGFhf9gs4xhRrGgeSegaY47wJ7dbub9LBWnFdCSbUS3zPFv1EQucT",
			  "base58"
			],
			"version": "legacy"
		},
		"id": 1
	}`

	rawResponse := struct {
		Result rpc.TransactionWithMeta `json:"result"`
	}{}
	err := json.Unmarshal([]byte(cancelTxJSON), &rawResponse)
	require.NoError(t, err)

	// --- act ---
	events, err := ParseTimelockEvents(&rawResponse.Result)
	require.NoError(t, err)

	// --- assert ---
	require.Len(t, events.Executed, 0)
	require.Len(t, events.Scheduled, 0)
	require.Len(t, events.Cancelled, 1)

	cancelEvent := events.Cancelled[0]
	expectedOperationKey := operationKeyFromHex("0xdeadc0ffeedeadc0ffeedeadc0ffeedeadc0ffeedeadc0ffeedeadc0ffee1234")
	require.Equal(t, expectedOperationKey, cancelEvent.ID)
}
