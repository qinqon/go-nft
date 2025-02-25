/*
 * This file is part of the go-nft project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2021 Red Hat, Inc.
 *
 */

package nft_test

import (
	"encoding/json"
	"fmt"
	"testing"

	assert "github.com/stretchr/testify/require"

	"github.com/networkplumbing/go-nft/nft"
	"github.com/networkplumbing/go-nft/nft/schema"
)

type ruleAction string

// Rule Actions
const (
	ruleADD    ruleAction = "add"
	ruleDELETE ruleAction = "delete"
)

func TestRule(t *testing.T) {
	testAddRuleWithMatchAndVerdict(t)
	testDeleteRule(t)

	testAddRuleWithRowExpression(t)

	testRuleLookup(t)

	testReadRuleWithNumericalExpression(t)
}

func testAddRuleWithRowExpression(t *testing.T) {
	const comment = "mycomment"

	table := nft.NewTable(tableName, nft.FamilyIP)
	chain := nft.NewRegularChain(table, chainName)

	t.Run("Add rule with a row expression, check serialization", func(t *testing.T) {
		statements, serializedStatements := matchWithRowExpression()
		rule := nft.NewRule(table, chain, statements, nil, nil, comment)

		config := nft.NewConfig()
		config.AddRule(rule)

		serializedConfig, err := config.ToJSON()
		assert.NoError(t, err)

		expectedConfig := buildSerializedConfig(ruleADD, serializedStatements, nil, comment)
		assert.Equal(t, string(expectedConfig), string(serializedConfig))
	})

	t.Run("Add rule with a row expression, check deserialization", func(t *testing.T) {
		statements, serializedStatements := matchWithRowExpression()

		serializedConfig := buildSerializedConfig(ruleADD, serializedStatements, nil, comment)

		var deserializedConfig nft.Config
		assert.NoError(t, json.Unmarshal(serializedConfig, &deserializedConfig))

		rule := nft.NewRule(table, chain, statements, nil, nil, comment)
		expectedConfig := nft.NewConfig()
		expectedConfig.AddRule(rule)

		assert.Equal(t, expectedConfig, &deserializedConfig)
	})
}

func testAddRuleWithMatchAndVerdict(t *testing.T) {
	const comment = "mycomment"

	table := nft.NewTable(tableName, nft.FamilyIP)
	chain := nft.NewRegularChain(table, chainName)

	t.Run("Add rule with match and verdict, check serialization", func(t *testing.T) {
		statements, serializedStatements := matchSrcIP4withReturnVerdict()
		rule := nft.NewRule(table, chain, statements, nil, nil, comment)

		config := nft.NewConfig()
		config.AddRule(rule)

		serializedConfig, err := config.ToJSON()
		assert.NoError(t, err)

		expectedConfig := buildSerializedConfig(ruleADD, serializedStatements, nil, comment)
		assert.Equal(t, string(expectedConfig), string(serializedConfig))
	})

	t.Run("Add rule with match and verdict, check deserialization", func(t *testing.T) {
		statements, serializedStatements := matchSrcIP4withReturnVerdict()

		serializedConfig := buildSerializedConfig(ruleADD, serializedStatements, nil, comment)

		var deserializedConfig nft.Config
		assert.NoError(t, json.Unmarshal(serializedConfig, &deserializedConfig))

		rule := nft.NewRule(table, chain, statements, nil, nil, comment)
		expectedConfig := nft.NewConfig()
		expectedConfig.AddRule(rule)

		assert.Equal(t, expectedConfig, &deserializedConfig)
	})
}

func testDeleteRule(t *testing.T) {
	table := nft.NewTable(tableName, nft.FamilyIP)
	chain := nft.NewRegularChain(table, chainName)

	t.Run("Delete rule", func(t *testing.T) {
		handleID := 100
		rule := nft.NewRule(table, chain, nil, &handleID, nil, "")

		config := nft.NewConfig()
		config.DeleteRule(rule)

		serializedConfig, err := config.ToJSON()
		assert.NoError(t, err)

		expectedConfig := buildSerializedConfig(ruleDELETE, "", &handleID, "")
		assert.Equal(t, string(expectedConfig), string(serializedConfig))
	})
}

func buildSerializedConfig(action ruleAction, serializedStatements string, handle *int, comment string) []byte {
	ruleArgs := fmt.Sprintf(`"family":%q,"table":%q,"chain":%q`, nft.FamilyIP, tableName, chainName)
	if serializedStatements != "" {
		ruleArgs += "," + serializedStatements
	}
	if handle != nil {
		ruleArgs += fmt.Sprintf(`,"handle":%d`, *handle)
	}
	if comment != "" {
		ruleArgs += fmt.Sprintf(`,"comment":%q`, comment)
	}

	var config string
	if action == ruleADD {
		config = fmt.Sprintf(`{"nftables":[{"rule":{%s}}]}`, ruleArgs)
	} else {
		config = fmt.Sprintf(`{"nftables":[{%q:{"rule":{%s}}}]}`, action, ruleArgs)
	}
	return []byte(config)
}

func matchSrcIP4withReturnVerdict() ([]schema.Statement, string) {
	ipAddress := "10.10.10.10"
	matchSrcIP4 := schema.Statement{
		Match: &schema.Match{
			Op: schema.OperEQ,
			Left: schema.Expression{
				Payload: &schema.Payload{
					Protocol: schema.PayloadProtocolIP4,
					Field:    schema.PayloadFieldIPSAddr,
				},
			},
			Right: schema.Expression{String: &ipAddress},
		},
	}

	verdict := schema.Statement{}
	verdict.Return = true

	statements := []schema.Statement{matchSrcIP4, verdict}

	expectedMatch := fmt.Sprintf(
		`"match":{"op":"==","left":{"payload":{"protocol":"ip","field":"saddr"}},"right":%q}`, ipAddress,
	)
	expectedVerdict := `"return":null`
	serializedStatements := fmt.Sprintf(`"expr":[{%s},{%s}]`, expectedMatch, expectedVerdict)

	return statements, serializedStatements
}

func matchWithRowExpression() ([]schema.Statement, string) {
	stringExpression := "string-expression"
	rowExpression := `{"foo":"boo"}`
	match := schema.Statement{
		Match: &schema.Match{
			Op:    schema.OperEQ,
			Left:  schema.Expression{RowData: json.RawMessage(rowExpression)},
			Right: schema.Expression{String: &stringExpression},
		},
	}

	statements := []schema.Statement{match}

	expectedMatch := fmt.Sprintf(`"match":{"op":"==","left":%s,"right":%q}`, rowExpression, stringExpression)
	serializedStatements := fmt.Sprintf(`"expr":[{%s}]`, expectedMatch)

	return statements, serializedStatements
}

func testRuleLookup(t *testing.T) {
	config := nft.NewConfig()
	table_br := nft.NewTable("table-br", nft.FamilyBridge)
	config.AddTable(table_br)

	chainRegular := nft.NewRegularChain(table_br, "chain-regular")
	config.AddChain(chainRegular)

	ruleSimple := nft.NewRule(table_br, chainRegular, nil, nil, nil, "comment123")
	config.AddRule(ruleSimple)

	ruleWithStatement := nft.NewRule(table_br, chainRegular, []schema.Statement{{}}, nil, nil, "comment456")
	ruleWithStatement.Expr[0].Drop = true
	config.AddRule(ruleWithStatement)

	handle := 10
	index := 1
	ruleWithAllParams := nft.NewRule(table_br, chainRegular, []schema.Statement{{}, {}}, &handle, &index, "comment789")
	config.AddRule(ruleWithAllParams)

	t.Run("Lookup an existing rule by table, chain and comment", func(t *testing.T) {
		rules := config.LookupRule(ruleSimple)
		assert.Len(t, rules, 1)
		assert.Equal(t, ruleSimple, rules[0])
	})

	t.Run("Lookup an existing rule by table, chain, statement and comment", func(t *testing.T) {
		rules := config.LookupRule(ruleWithStatement)
		assert.Len(t, rules, 1)
		assert.Equal(t, ruleWithStatement, rules[0])
	})

	t.Run("Lookup an existing rule by all (root) parameters", func(t *testing.T) {
		rules := config.LookupRule(ruleWithAllParams)
		assert.Len(t, rules, 1)
		assert.Equal(t, ruleWithAllParams, rules[0])
	})

	t.Run("Lookup a missing rule (comment not matching)", func(t *testing.T) {
		rule := nft.NewRule(table_br, chainRegular, nil, nil, nil, "comment-missing")
		assert.Empty(t, config.LookupRule(rule))
	})

	t.Run("Lookup a missing rule (statement content not matching)", func(t *testing.T) {
		rule := nft.NewRule(table_br, chainRegular, []schema.Statement{{}}, nil, nil, "comment456")
		rule.Expr[0].Drop = false
		rule.Expr[0].Return = true
		assert.Empty(t, config.LookupRule(rule))
	})

	t.Run("Lookup a missing rule (statements count not matching)", func(t *testing.T) {
		rule := nft.NewRule(table_br, chainRegular, []schema.Statement{{}, {}}, nil, nil, "comment456")
		rule.Expr[0].Drop = true
		assert.Empty(t, config.LookupRule(rule))
	})

	t.Run("Lookup a missing rule (handle not matching)", func(t *testing.T) {
		changedHandle := 99
		rule := nft.NewRule(table_br, chainRegular, []schema.Statement{{}, {}}, &changedHandle, &index, "comment789")
		assert.Empty(t, config.LookupRule(rule))
	})
}

func testReadRuleWithNumericalExpression(t *testing.T) {
	t.Run("Read rule with numerical expression", func(t *testing.T) {
		c := nft.NewConfig()
		assert.NoError(t, c.FromJSON([]byte(`
		{"nftables":[{"rule":{
		   "expr":[{"match":{"op":"==","left":"foo","right":12345}}]
		}}]}
		`)))
	})
}
