// Domain types aligned with the OpenAPI TransactionCreate schema

export type TransactionType = 'income' | 'expense';

export interface TransactionCreateDTO {
  type: TransactionType;
  amount: number;
  currency: string;
  description: string;
  date: string; // ISO date YYYY-MM-DD
}

export interface TransactionResponse {
  id: string;
}

export interface ApiError {
  code: string;
  message: string;
}
