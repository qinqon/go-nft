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
	"testing"

	assert "github.com/stretchr/testify/require"

	"github.com/networkplumbing/go-nft/nft"
	"github.com/networkplumbing/go-nft/nft/schema"
)

type tableActionFunc func(*nft.Config, *schema.Table)

const tableName = "test-table"

func TestTable(t *testing.T) {
	testTableActions(t)
	testTableLookup(t)
}

func testTableActions(t *testing.T) {
	actions := map[nft.TableAction]tableActionFunc{
		nft.TableADD:    func(c *nft.Config, t *schema.Table) { c.AddTable(t) },
		nft.TableDELETE: func(c *nft.Config, t *schema.Table) { c.DeleteTable(t) },
		nft.TableFLUSH:  func(c *nft.Config, t *schema.Table) { c.FlushTable(t) },
	}
	families := []nft.AddressFamily{
		nft.FamilyIP,
		nft.FamilyIP6,
		nft.FamilyINET,
		nft.FamilyBridge,
		nft.FamilyARP,
		nft.FamilyNETDEV,
	}
	for actionName, actionFunc := range actions {
		for _, family := range families {
			testTableAction(t, actionName, actionFunc, family)
		}
	}
}

func testTableAction(t *testing.T, actionName nft.TableAction, actionFunc tableActionFunc, family nft.AddressFamily) {
	testName := fmt.Sprintf("%s %s table", actionName, family)
	t.Run(testName, func(t *testing.T) {
		table := nft.NewTable(tableName, family)
		config := nft.NewConfig()
		actionFunc(config, table)

		serializedConfig, err := config.ToJSON()
		assert.NoError(t, err)

		var expected []byte
		if actionName == nft.TableADD {
			expected = []byte(fmt.Sprintf(`{"nftables":[{"table":{"family":%q,"name":%q}}]}`, family, tableName))
		} else {
			expected = []byte(fmt.Sprintf(`{"nftables":[{%q:{"table":{"family":%q,"name":%q}}}]}`, actionName, family, tableName))
		}
		assert.Equal(t, string(expected), string(serializedConfig))
	})
}

func testTableLookup(t *testing.T) {
	config := nft.NewConfig()
	config.AddTable(nft.NewTable("table-ip", nft.FamilyIP))
	config.AddTable(nft.NewTable("table-ip", nft.FamilyIP6))
	table_br := nft.NewTable("table-br", nft.FamilyBridge)
	config.AddTable(table_br)

	config.AddChain(nft.NewRegularChain(table_br, "chain-br"))

	t.Run("Lookup an existing table", func(t *testing.T) {
		table := config.LookupTable(table_br)
		assert.Equal(t, *table_br, *table)
	})

	t.Run("Lookup a missing table", func(t *testing.T) {
		table := config.LookupTable(nft.NewTable("table-na", nft.FamilyBridge))
		assert.Nil(t, table)
	})
}
