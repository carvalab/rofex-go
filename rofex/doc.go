// Package rofex is a Go client for the Primary trading API, the API used by
// MATBA ROFEX and brokers such as Eco Valores to access the Argentine futures,
// options and BYMA (stocks, bonds, CEDEARs) markets. It is the Go counterpart
// of pyRofex.
//
// It covers the REST and WebSocket APIs:
//
//   - Market data: real-time snapshots, WebSocket streaming, historical trades and candles
//   - Orders: send, cancel, replace, status and execution report streaming
//   - Accounts and risk: positions, detailed positions and account reports
//   - Reference data: segments and instruments
//
// Authentication is automatic: [NewPasswordAuth] logs in on the first request
// and refreshes the token when it expires.
//
// # Environments
//
// [model.EnvironmentRemarket] (default) targets the free reMarkets sandbox
// (https://remarkets.primary.ventures). For production use
// [model.EnvironmentLive] and set your broker URLs with [WithBaseURL] and
// [WithWSURL].
//
// See the examples below and docs/primary-api.md for the API reference.
package rofex
