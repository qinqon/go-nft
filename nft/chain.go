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

package nft

import (
	"github.com/networkplumbing/go-nft/nft/schema"
)

type ChainType string
type ChainHook string
type ChainPolicy string

// Chain Types
const (
	TypeFilter ChainType = schema.TypeFilter
	TypeNAT    ChainType = schema.TypeNAT
	TypeRoute  ChainType = schema.TypeRoute
)

// Chain Hooks
const (
	HookPreRouting  ChainHook = schema.HookPreRouting
	HookInput       ChainHook = schema.HookInput
	HookOutput      ChainHook = schema.HookOutput
	HookForward     ChainHook = schema.HookForward
	HookPostRouting ChainHook = schema.HookPostRouting
	HookIngress     ChainHook = schema.HookIngress
)

// Chain Policies
const (
	PolicyAccept ChainPolicy = schema.PolicyAccept
	PolicyDrop   ChainPolicy = schema.PolicyDrop
)

// NewRegularChain returns a new schema chain structure for a regular chain.
func NewRegularChain(table *schema.Table, name string) *schema.Chain {
	return NewChain(table, name, nil, nil, nil, nil)
}

// NewChain returns a new schema chain structure for a base chain.
// For base chains, all arguments are required except the policy.
// Missing arguments will cause an error once the config is applied.
func NewChain(table *schema.Table, name string, ctype *ChainType, hook *ChainHook, prio *int, policy *ChainPolicy) *schema.Chain {
	c := &schema.Chain{
		Family: table.Family,
		Table:  table.Name,
		Name:   name,
	}

	if ctype != nil {
		c.Type = string(*ctype)
	}
	if hook != nil {
		c.Hook = string(*hook)
	}
	if prio != nil {
		c.Prio = prio
	}
	if policy != nil {
		c.Policy = string(*policy)
	}

	return c
}

// AddChain appends the given chain to the nftable config.
// The chain is added without an explicit action (`add`).
// Adding multiple times the same chain has no affect when the config is applied.
func (c *Config) AddChain(chain *schema.Chain) {
	nftable := schema.Nftable{Chain: chain}
	c.Nftables = append(c.Nftables, nftable)
}

// DeleteChain appends a given chain to the nftable config
// with the `delete` action.
// Attempting to delete a non-existing chain, results with a failure when the config is applied.
// The chain must not contain any rules or be used as a jump target.
func (c *Config) DeleteChain(chain *schema.Chain) {
	nftable := schema.Nftable{Delete: &schema.Objects{Chain: chain}}
	c.Nftables = append(c.Nftables, nftable)
}

// FlushChain appends a given chain to the nftable config
// with the `flush` action.
// All rules under the chain are removed (when applied).
// Attempting to flush a non-existing chain, results with a failure when the config is applied.
func (c *Config) FlushChain(chain *schema.Chain) {
	nftable := schema.Nftable{Flush: &schema.Objects{Chain: chain}}
	c.Nftables = append(c.Nftables, nftable)
}

// LookupChain searches the configuration for a matching chain and returns it.
// The chain is matched first by the table and chain name.
// Other matching fields are optional (for matching base chains).
// Mutating the returned chain will result in mutating the configuration.
func (c *Config) LookupChain(toFind *schema.Chain) *schema.Chain {
	for _, nftable := range c.Nftables {
		if chain := nftable.Chain; chain != nil {
			match := chain.Table == toFind.Table && chain.Family == toFind.Family && chain.Name == toFind.Name
			if match {
				if t := toFind.Type; t != "" {
					match = match && chain.Type == t
				}
				if h := toFind.Hook; h != "" {
					match = match && chain.Hook == h
				}
				if p := toFind.Prio; p != nil {
					match = match && chain.Prio != nil && *chain.Prio == *p
				}
				if p := toFind.Policy; p != "" {
					match = match && chain.Policy == p
				}
				if match {
					return chain
				}
			}
		}
	}
	return nil
}
