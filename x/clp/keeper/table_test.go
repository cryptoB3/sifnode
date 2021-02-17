package keeper

import (
	"bytes"
	"encoding/json"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"strings"
	"testing"
)

func TestCalculatePoolUnits(t *testing.T) {
	type TestCase struct {
		NativeAdded      string `json:"r"`
		ExternalAdded    string `json:"a"`
		NativeBalance    string `json:"R"`
		ExternalBalance  string `json:"A"`
		PoolUnitsBalance string `json:"P"`
		Expected         string `json:"expected"`
	}
	type Test struct {
		TestType []TestCase `json:"PoolUnits"`
	}
	file, err := ioutil.ReadFile("../../../test/test-tables/pool_units.json")
	assert.NoError(t, err)
	file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf"))
	var test Test
	err = json.Unmarshal(file, &test)
	assert.NoError(t, err)
	testcases := test.TestType
	errCount := 0
	for _, test := range testcases {
		_, stakeUnits, _ := CalculatePoolUnits(
			"cusdt",
			sdk.NewUintFromString(test.PoolUnitsBalance),
			sdk.NewUintFromString(test.NativeBalance),
			sdk.NewUintFromString(test.ExternalBalance),
			sdk.NewUintFromString(test.NativeAdded),
			sdk.NewUintFromString(test.ExternalAdded),
		)
		if test.Expected != "0" && !stakeUnits.Equal(sdk.NewUintFromString(test.Expected)) {
			errCount++
			fmt.Printf("Got %s , Expected %s \n", stakeUnits, test.Expected)
		}

	}
	fmt.Printf("TestCalculatePoolUnits \nTotal/Failed: %d/%d\n", len(testcases), errCount)
}

func TestCalculateSwapResult(t *testing.T) {
	type TestCase struct {
		Xx       string `json:"x"`
		X        string `json:"X"`
		Y        string `json:"Y"`
		Expected string `json:"expected"`
	}

	type Test struct {
		TestType []TestCase `json:"SingleSwapResult"`
	}
	file, err := ioutil.ReadFile("../../../test/test-tables/singleswap_result.json")
	assert.NoError(t, err)

	file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf"))
	var test Test
	err = json.Unmarshal(file, &test)
	assert.NoError(t, err)

	testcases := test.TestType
	errCount := 0
	for _, test := range testcases {
		Yy, _ := calcSwapResult("cusdt",
			true,
			sdk.NewUintFromString(test.X),
			sdk.NewUintFromString(test.Xx),
			sdk.NewUintFromString(test.Y))
		if test.Expected != "0" && !Yy.Equal(sdk.NewUintFromString(test.Expected)) {
			errCount++
			fmt.Printf("Got %s , Expected %s \n", Yy, strings.Split(test.Expected, ".")[0])
		}
	}
	fmt.Printf("TestCalculateSwapResult \nTotal/Failed: %d/%d\n", len(testcases), errCount)
}

func TestCalculateSwapLiquidityFee(t *testing.T) {
	type TestCase struct {
		Xx       string `json:"x"`
		X        string `json:"X"`
		Y        string `json:"Y"`
		Expected string `json:"expected"`
	}

	type Test struct {
		TestType []TestCase `json:"SingleSwapLiquidityFee"`
	}
	file, err := ioutil.ReadFile("../../../test/test-tables/singleswap_liquidityfees.json")
	assert.NoError(t, err)

	file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf"))
	var test Test
	err = json.Unmarshal(file, &test)
	assert.NoError(t, err)

	testcases := test.TestType
	errCount := 0
	for _, test := range testcases {
		Yy, _ := calcLiquidityFee("cusdt",
		    true,
			sdk.NewUintFromString(test.X),
			sdk.NewUintFromString(test.Xx),
			sdk.NewUintFromString(test.Y))
		if test.Expected != "0" && !Yy.Equal(sdk.NewUintFromString(test.Expected)) {
			errCount++
			fmt.Printf("Got %s , Expected %s \n", Yy, strings.Split(test.Expected, ".")[0])
		}
	}
	fmt.Printf("TestCalculateSwapLiquidityFee \nTotal/Failed: %d/%d\n", len(testcases), errCount)
}

func TestCalculateDoubleSwapResult(t *testing.T) {
	type TestCase struct {
		Ax       string `json:"ax"`
		AX       string `json:"aX"`
		AY       string `json:"aY"`
		BX       string `json:"bX"`
		BY       string `json:"bY"`
		Expected string `json:"expected"`
	}

	type Test struct {
		TestType []TestCase `json:"DoubleSwap"`
	}
	file, err := ioutil.ReadFile("../../../test/test-tables/doubleswap_result.json")
	assert.NoError(t, err)

	file = bytes.TrimPrefix(file, []byte("\xef\xbb\xbf"))
	var test Test
	err = json.Unmarshal(file, &test)
	assert.NoError(t, err)

	testcases := test.TestType
	errCount := 0
	for _, test := range testcases {
		Ay, _ := calcSwapResult("cusdt",
			true,
			sdk.NewUintFromString(test.AX),
			sdk.NewUintFromString(test.Ax),
			sdk.NewUintFromString(test.AY))

		By, _ := calcSwapResult("cusdt",
			true,
			sdk.NewUintFromString(test.BX),
			Ay,
			sdk.NewUintFromString(test.BY))

		if test.Expected != "0" && !By.Equal(sdk.NewUintFromString(test.Expected)) {
			errCount++
			fmt.Printf("TestCalculateDoubleSwapResult: Was %s, Got %s , Expected %s \n %v \n", Ay, By, strings.Split(test.Expected, ".")[0], test)
		}
	}
	fmt.Printf("\nTotal/Failed: %d/%d \n", len(testcases), errCount)
}
