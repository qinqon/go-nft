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
	"fmt"
	"strings"
	"testing"

	assert "github.com/stretchr/testify/require"

	"github.com/networkplumbing/go-nft/nft"
	"github.com/networkplumbing/go-nft/nft/schema"
)

type chainAction string

type chainActionFunc func(*nft.Config, *schema.Chain)

// Chain Actions
const (
	chainADD    chainAction = "add"
	chainDELETE chainAction = "delete"
	chainFLUSH  chainAction = "flush"
)

const chainName = "test-chain"

func TestChain(t *testing.T) {
	testAddBaseChains(t)
	// Removal of base-chains is identical to the removal of regular-chains.
	// Therefore, such scenarios are evaluated through the regular-chains actions
	testRegularChainsActions(t)

	testChainLookup(t)
}

func testAddBaseChains(t *testing.T) {
	types := []nft.ChainType{
		nft.TypeFilter,
		nft.TypeNAT,
		nft.TypeRoute,
	}
	hooks := []nft.ChainHook{
		nft.HookPreRouting,
		nft.HookInput,
		nft.HookOutput,
		nft.HookForward,
		nft.HookPostRouting,
		nft.HookIngress,
	}
	policies := []nft.ChainPolicy{
		nft.PolicyAccept,
		nft.PolicyDrop,
		"",
	}

	table := nft.NewTable(tableName, nft.FamilyIP)

	for _, ctype := range types {
		for _, hook := range hooks {
			for _, policy := range policies {
				testName := createChainTestName(chainADD, ctype, hook, policy)

				t.Run(testName, func(t *testing.T) {
					priority := 100
					chain := nft.NewChain(table, chainName, &ctype, &hook, &priority, &policy)
					config := nft.NewConfig()
					config.AddChain(chain)

					serializedConfig, err := config.ToJSON()
					assert.NoError(t, err)

					chainArgs := fmt.Sprintf(
						`"family":%q,"table":%q,"name":%q,"type":%q,"hook":%q,"prio":%d`,
						table.Family, table.Name, chainName, ctype, hook, priority,
					)
					if policy != "" {
						chainArgs += fmt.Sprintf(`,"policy":%q`, policy)
					}
					expected := []byte(fmt.Sprintf(`{"nftables":[{"chain":{%s}}]}`, chainArgs))
					assert.Equal(t, string(expected), string(serializedConfig))
				})
			}
		}
	}
}

func testRegularChainsActions(t *testing.T) {
	actions := map[chainAction]chainActionFunc{
		chainADD:    func(c *nft.Config, chain *schema.Chain) { c.AddChain(chain) },
		chainDELETE: func(c *nft.Config, chain *schema.Chain) { c.DeleteChain(chain) },
		chainFLUSH:  func(c *nft.Config, chain *schema.Chain) { c.FlushChain(chain) },
	}

	table := nft.NewTable(tableName, nft.FamilyIP)
	chain := nft.NewRegularChain(table, chainName)

	for action, actionFunc := range actions {
		testName := createChainTestName(action, "", "", "")

		t.Run(testName, func(t *testing.T) {
			config := nft.NewConfig()
			actionFunc(config, chain)

			serializedConfig, err := config.ToJSON()
			assert.NoError(t, err)

			chainArgs := fmt.Sprintf(`"family":%q,"table":%q,"name":%q`, table.Family, table.Name, chainName)
			var expected []byte
			if action == chainADD {
				expected = []byte(fmt.Sprintf(`{"nftables":[{"chain":{%s}}]}`, chainArgs))
			} else {
				expected = []byte(fmt.Sprintf(`{"nftables":[{%q:{"chain":{%s}}}]}`, action, chainArgs))
			}

			assert.Equal(t, string(expected), string(serializedConfig))
		})
	}
}

func createChainTestName(action chainAction, ctype nft.ChainType, hook nft.ChainHook, policy nft.ChainPolicy) string {
	args := []string{string(action)}
	if ctype != "" {
		args = append(args, string(ctype))
	}
	if hook != "" {
		args = append(args, string(hook))
	}
	if policy != "" {
		args = append(args, string(policy))
	}

	return strings.Join(args, " ")
}

func testChainLookup(t *testing.T) {
	config := nft.NewConfig()
	table_br := nft.NewTable("table-br", nft.FamilyBridge)
	config.AddTable(table_br)

	chainRegular := nft.NewRegularChain(table_br, "chain-regular")
	config.AddChain(chainRegular)

	ctype, hook, prio, policy := nft.TypeFilter, nft.HookPreRouting, 100, nft.PolicyAccept
	chainBase := nft.NewChain(table_br, "chain-base", &ctype, &hook, &prio, &policy)
	config.AddChain(chainBase)

	t.Run("Lookup an existing regular chain", func(t *testing.T) {
		chain := config.LookupChain(chainRegular)
		assert.Equal(t, chainRegular, chain)
	})

	t.Run("Lookup an existing base chain", func(t *testing.T) {
		chain := config.LookupChain(chainBase)
		assert.Equal(t, chainBase, chain)
	})

	t.Run("Lookup a missing regular chain", func(t *testing.T) {
		chain := nft.NewRegularChain(table_br, "chain-na")
		assert.Nil(t, config.LookupChain(chain))
	})

	t.Run("Lookup a missing base chain", func(t *testing.T) {
		inputHook := nft.HookInput
		chain := nft.NewChain(table_br, "chain-base", &ctype, &inputHook, &prio, &policy)
		assert.Nil(t, config.LookupChain(chain))
	})
}
