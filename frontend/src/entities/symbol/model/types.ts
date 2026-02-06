export interface SymbolSpec {
  symbol: string;
  base_currency: string;
  quote_currency: string;
  min_quantity: string;
  max_quantity: string;
  quantity_step: string;
  min_leverage: number;
  max_leverage: number;
  maintenance_rate: string;
}
