import type { TransactionCreateDTO, TransactionResponse } from '../types';

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8080/api/v1';

export async function createTransaction(
  payload: TransactionCreateDTO
): Promise<TransactionResponse> {
  const res = await fetch(`${API_BASE}/transactions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });

  if (!res.ok) {
    const err = await res.json().catch(() => ({ code: 'unknown', message: 'Unknown error' }));
    throw new Error(err.message ?? 'Failed to create transaction');
  }

  return res.json() as Promise<TransactionResponse>;
}
