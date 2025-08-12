package rofex

import (
	"context"
	"fmt"

	"github.com/carvalab/rofex-go/rofex/model"
)

// Accounts lista las cuentas asociadas al usuario autenticado.
//
// Devuelve todas las cuentas disponibles para el usuario actual.
// Útil para determinar qué cuentas se pueden usar para trading.
func (c *Client) Accounts(ctx context.Context) (model.AccountsResponse, error) {
	return getTyped[model.AccountsResponse](ctx, c, pathAccounts)
}

// AccountPosition consulta las posiciones de una cuenta según Primary Risk API.
//
// Primary Risk API es la interfaz que permite comunicarse con el sistema RIMA (Risk Manager).
// Este método devuelve las posiciones actuales para una cuenta específica.
//
// La respuesta incluye para cada posición:
//   - Instrumento y símbolo de trading
//   - Cantidades y precios de compra/venta (buySize, sellSize, buyPrice, sellPrice)
//   - Diferencias diarias y totales (totalDailyDiff, totalDiff)
//   - Precios originales de compra/venta
//
// Referencia: docs/primary-api.md - "Consultar las posiciones de una cuenta"
func (c *Client) AccountPosition(ctx context.Context, account string) (model.AccountPositionResponse, error) {
	if account == "" {
		return model.AccountPositionResponse{}, &ValidationError{Field: "account", Msg: "required"}
	}
	path := fmt.Sprintf(pathAccountPos, account)
	return getTyped[model.AccountPositionResponse](ctx, c, path)
}

// DetailedPosition consulta el detalle de posiciones según Primary Risk API.
//
// Devuelve información detallada y agrupada por tipo de contrato incluyendo:
//   - Posiciones iniciales, operadas y actuales por instrumento
//   - Diferencias diarias detalladas
//   - Agrupación por tipo de contrato (FUTURE, FUTURE_OPTION_CALL, etc.)
//   - Fecha de último cálculo
//
// Referencia: docs/primary-api.md - "Consultar detalle de posiciones"
func (c *Client) DetailedPosition(ctx context.Context, account string) (model.DetailedPositionResponse, error) {
	if account == "" {
		return model.DetailedPositionResponse{}, &ValidationError{Field: "account", Msg: "required"}
	}
	path := fmt.Sprintf(pathDetailedPos, account)
	return getTyped[model.DetailedPositionResponse](ctx, c, path)
}

// AccountReport consulta el reporte de cuenta según Primary Risk API.
//
// Este método devuelve un reporte completo de la cuenta incluyendo:
//   - Balance de efectivo y márgenes por moneda
//   - Disponible para operar (cash, movimientos, crédito)
//   - Portfolio y márgenes de órdenes
//   - Diferencias diarias y efectivo actual
//   - Fecha de liquidación y cálculos de riesgo
//
// ⚠️ Rate Limit: Máximo 1 request cada 5 segundos para reportes según documentación oficial.
//
// El reporte soporta multimoneda y proporciona información detallada
// del estado financiero de la cuenta.
//
// Referencia: docs/primary-api.md - "Consultar reporte de cuenta"
func (c *Client) AccountReport(ctx context.Context, account string) (model.AccountReportResponse, error) {
	if account == "" {
		return model.AccountReportResponse{}, &ValidationError{Field: "account", Msg: "required"}
	}
	path := fmt.Sprintf(pathAccountReport, account)
	return getTyped[model.AccountReportResponse](ctx, c, path)
}
