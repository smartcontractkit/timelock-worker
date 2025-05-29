package timelock

import (
	"encoding/json"
	"testing"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/stretchr/testify/require"
)

func TestParseTimelockEvents_CallExecutedTx(t *testing.T) {
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
    "transaction": {
      "message": {
        "accountKeys": [
          "7oZnxiocDK1aa9XAQC3CZ1VHKFkKwLuwRK8NddhU3FT2",
          "2KVM8NpDNHpUDxPDW97YdAQZRENVCL4gDuyMvWtqUbj4",
          "G88p8xbxwG6D5ZRbx7S31sdPPYR2UnRs1PhhE5iAL1ny",
          "DDT28LawDCqFwPqjp8t35UzheUVWk24qZJN6gFHKwsnf",
          "DJnQbX4PrA4854CpsA5hDkizx1drAccQbHE8jZrgRhAk",
          "35u11sTYbcen34onkPHVEaekkJ4ua4k1SqkXV2x7bEPy",
          "HpYr8NAom2tEpdp5mP1pi5RYX6bfd5Mk3emH2CgmzG6D",
          "offqSMQWgQud6WJz694LRzkeN5kMYpCHTpXQr3Rkcjm",
          "DoajfR5tK24xVw51fWcawUZWhAXD8yrBJVacc13neVQA",
          "ComputeBudget111111111111111111111111111111"
        ],
        "header": {
          "numReadonlySignedAccounts": 0,
          "numReadonlyUnsignedAccounts": 7,
          "numRequiredSignatures": 1
        },
        "instructions": [
          {
            "accounts": [
              1,
              3,
              4,
              5,
              6,
              0,
              7,
              2,
              5
            ],
            "data": "2fRCKgyB2eYrGZR5XdcvfK3ckk36FUPhL2WTTBmCCsWdTX22WuTj4HA7x6djCELSmy1mCwXZGmneFqntv34DvHL5LuoBQASjphf",
            "programIdIndex": 8,
            "stackHeight": null
          },
          {
            "accounts": [],
            "data": "K1FDJ7",
            "programIdIndex": 9,
            "stackHeight": null
          }
        ],
        "recentBlockhash": "JDSXdiPBLNM7eB4tq9fmAi5PjdvhmskTUwAYbZp3PkKC"
      },
      "signatures": [
        "4dD688oj9n788AXYHnUWzKeEyZ3fV32DRrPozTQ4tXrNYVYDqxxhhPQK8axxt6gCD7TF5sFmkeKbHnFYc5kkZ9DU"
      ]
    },
    "version": "legacy"
  },
  "id": 1
}`
	var rawResponse struct {
		Result rpc.TransactionWithMeta `json:"result"`
	}
	err := json.Unmarshal([]byte(executeBatchTxJSON), &rawResponse)
	require.NoError(t, err, "should unmarshal JSON into rpc.TransactionWithMeta")
	evs, err := ParseTimelockEvents(testLogger, &rawResponse.Result)
	require.NoError(t, err)
	// Expect exactly one CallScheduled and one CallExecuted
	require.Len(t, evs.Executed, 1, "should find one CallExecuted event")
	require.Len(t, evs.Scheduled, 0, "should find 0 Scheduled event")
	require.Len(t, evs.Cancelled, 0, "should find 0 Cancelled event")
	// Check the CallExecuted event
	evt := evs.Executed[0]
	require.Equal(t, "offqSMQWgQud6WJz694LRzkeN5kMYpCHTpXQr3Rkcjm", evt.Target.String(), "should match target public key")
	require.Equal(t, "\xac\x17+\r\xee\xd5U\x96", string(evt.Data), "should match event data")
	require.Equal(t, uint64(0), evt.Index, "should match event index")
	require.Equal(t,
		[32]uint8{0x9b, 0x33, 0x3d, 0xdb, 0x26, 0xb6, 0x63, 0x7, 0x66, 0xb0, 0x65, 0xbb, 0xf, 0x20, 0x11, 0x3, 0x42, 0x2, 0x34, 0x84, 0x73, 0x93, 0x3f, 0x72, 0xfb, 0xf5, 0xea, 0x46, 0xeb, 0x5d, 0x51, 0x52},
		evt.ID,
		"should match event ID")
}

func TestParseTimelockEvents_CallScheduledTx(t *testing.T) {
	const scheduleBatchTxJSON = `
	{
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
    "transaction": {
      "message": {
        "accountKeys": [
          "7oZnxiocDK1aa9XAQC3CZ1VHKFkKwLuwRK8NddhU3FT2",
          "HznXSrT5dkweA5Ap86kYCabkU9S2Hf2Q5Qvno6Wkq9PH",
          "9nLpt2VKZBBHnj37tL4rfH8oAw4XygZRM9mAVMkHVVWN",
          "ZMyBiNGpJYK6BP5ubRSDJD7SimPQULp9zbgA27E968S",
          "9K5QmiFUayo3Hsmjt47VgJ8HeWuVgrUToAk9Huw5XmLk",
          "GMqkrQkqVFuk8xzwNLnSibWZQwpq7LvUEt5f1RdR3P68",
          "DoajfR5tK24xVw51fWcawUZWhAXD8yrBJVacc13neVQA",
          "DJnQbX4PrA4854CpsA5hDkizx1drAccQbHE8jZrgRhAk",
          "FTDusxFg9NmmFGRg5jfA9nHCiCpZ7dJktawfRBcUBhq",
          "5vNJx78mz7KVMjhuipyr9jKBKcMrKYGdjGkgE4LUmjKk",
          "ComputeBudget111111111111111111111111111111"
        ],
        "header": {
          "numReadonlySignedAccounts": 0,
          "numReadonlyUnsignedAccounts": 6,
          "numRequiredSignatures": 1
        },
        "instructions": [
          {
            "accounts": [
              1,
              5,
              2,
              6,
              4,
              0,
              3,
              7,
              8,
              4
            ],
            "data": "4JCzHhKSdCxjsntSMTMfneF3BKDQP1nVUTfuihwp5M4NtzjdSkWEGqF8muY9A9W2u5P9a36oNpb4rMhwLQCvpHw4kB8hfm9Sc1Zx5nwsPNJjvv6HLBGXvU8Dbi8UaSoSzBSFDrYcrGy1eqXcZJg8dFEh9hpnQw5tjZN884SD9D4M8S6ixMneF8F78FF7dFQPizAVWxzXf5Y3ydwwKvM5zjhnngmQu65sJwqbAekWKjhfZd1MdK2pNiu6tXNrVbHFYG6ML6fncDzBs1XoWsJgKwHreoWkaK6PgU5DcwSVQWgEuvf4VTKrYyW9zYbQ9PjWXNaUeuLLuw7MiVpURJS885DuHJCdUiSx8GbbV6zmRMm5DjSAB9Eh",
            "programIdIndex": 9,
            "stackHeight": null
          },
          {
            "accounts": [],
            "data": "K1FDJ7",
            "programIdIndex": 10,
            "stackHeight": null
          }
        ],
        "recentBlockhash": "G8ieKkfmP1HpXNmp2UktdjzTtbBpqCenFPYkmLQZFDoy"
      },
      "signatures": [
        "2G3nDopUPP1b4B8woZWVNME2Qo3i9eyAB7n1B2QLHHgM8gPb1dTK3iLkCqFyktgK9Jm4mU3Ny9aBahrcHUSzyTjg"
      ]
    },
    "version": "legacy"
  },
  "id": 1
}`
	var rawResponse struct {
		Result rpc.TransactionWithMeta `json:"result"`
	}
	err := json.Unmarshal([]byte(scheduleBatchTxJSON), &rawResponse)
	require.NoError(t, err, "should unmarshal JSON into rpc.TransactionWithMeta")
	evs, err := ParseTimelockEvents(testLogger, &rawResponse.Result)
	require.NoError(t, err)
	// Expect exactly one CallScheduled and one CallExecuted
	require.Len(t, evs.Executed, 0, "should find one CallExecuted event")
	require.Len(t, evs.Scheduled, 6, "should find 6 Scheduled events")
	require.Len(t, evs.Cancelled, 0, "should find 0 Cancelled event")
	// Check the Scheduled event
	evt := evs.Scheduled[0]
	require.Equal(t, "Ccip842gzYHhvdDkSyi2YVCoAWPbYJoApMFzSxQroE9C", evt.Target.String(), "should match target public key")
	require.Equal(t, "\xda%\x8bk\x8e\xe43\xdb\x1e\xf8~t\xbeH\xb1D\xe8\xeb\xe0\x03\x16\xe8\xf9/\x8a\xa8\r\xa7\x9b\xf8\r\xae\xfe\x87\xd1\xff|\x05mX", string(evt.Data), "should match event data")
	require.Equal(t, uint64(0), evt.Index, "should match event index")
	require.Equal(t,
		[32]uint8{0xc8, 0x5e, 0x96, 0xb7, 0xc5, 0xdf, 0xd3, 0x70, 0xe3, 0xd5, 0x11, 0x2f, 0xf0, 0x59, 0xd5, 0x5e, 0x95, 0xcc, 0x29, 0x95, 0x33, 0x6d, 0xf1, 0x3f, 0x37, 0x99, 0x36, 0xd4, 0x13, 0x35, 0x8a, 0x78},
		evt.ID,
		"should match event ID")
}

func TestParseTimelockEvents_CancelledEvent(t *testing.T) {
	const cancelTxJSON = `
{
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
      "postBalances": [1000000000, 200000000],
      "preBalances": [1000005000, 200000000],
      "postTokenBalances": [],
      "preTokenBalances": [],
      "rewards": [],
      "status": {
        "Ok": null
      }
    },
    "slot": 383999999,
    "transaction": {
      "message": {
        "accountKeys": [
          "Key111111111111111111111111111111111111111",
          "Key222222222222222222222222222222222222222"
        ],
        "header": {
          "numReadonlySignedAccounts": 0,
          "numReadonlyUnsignedAccounts": 1,
          "numRequiredSignatures": 1
        },
        "instructions": [],
        "recentBlockhash": "ABCDEFGH1234567890"
      },
      "signatures": [
        "sig111111111111111111111111111111111111111"
      ]
    },
    "version": "legacy"
  },
  "id": 1
}`

	var rawResponse struct {
		Result rpc.TransactionWithMeta `json:"result"`
	}
	err := json.Unmarshal([]byte(cancelTxJSON), &rawResponse)
	require.NoError(t, err)

	evs, err := ParseTimelockEvents(testLogger, &rawResponse.Result)
	require.NoError(t, err)

	require.Len(t, evs.Cancelled, 1, "should find 1 Cancelled event")

	cancelEvent := evs.Cancelled[0]
	expectedID := [32]byte{0xde, 0xad, 0xc0, 0xff, 0xee, 0xde, 0xad, 0xc0, 0xff, 0xee, 0xde, 0xad, 0xc0, 0xff, 0xee, 0xde, 0xad, 0xc0, 0xff, 0xee, 0xde, 0xad, 0xc0, 0xff, 0xee, 0xde, 0xad, 0xc0, 0xff, 0xee, 0x12, 0x34}

	require.Equal(t, expectedID, cancelEvent.ID, "should match cancelled event ID")
}
