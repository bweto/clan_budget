'use client';

import React, { useState, useId } from 'react';
import { createTransaction } from '../services/transactions';
import type { TransactionType } from '../types';
import styles from './QuickAddForm.module.css';

interface FormState {
  description: string;
  amount: string;
  type: TransactionType;
}

interface SubmitState {
  status: 'idle' | 'loading' | 'success' | 'error';
  message: string;
}

export function QuickAddForm() {
  const descId = useId();
  const amountId = useId();
  const typeId = useId();

  const [form, setForm] = useState<FormState>({
    description: '',
    amount: '',
    type: 'expense',
  });

  const [submit, setSubmit] = useState<SubmitState>({ status: 'idle', message: '' });

  const handleChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>
  ) => {
    const { name, value } = e.target;
    setForm(prev => ({ ...prev, [name]: value }));
    if (submit.status === 'error') setSubmit({ status: 'idle', message: '' });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!form.description.trim()) {
      setSubmit({ status: 'error', message: 'La descripción es requerida.' });
      return;
    }
    const amount = parseFloat(form.amount);
    if (isNaN(amount) || amount <= 0) {
      setSubmit({ status: 'error', message: 'El monto debe ser un número mayor a 0.' });
      return;
    }

    setSubmit({ status: 'loading', message: '' });

    try {
      await createTransaction({
        description: form.description.trim(),
        amount,
        type: form.type,
        currency: 'USD',
        date: new Date().toISOString().split('T')[0],
      });
      setSubmit({ status: 'success', message: '¡Gasto registrado exitosamente! 🎉' });
      setForm({ description: '', amount: '', type: 'expense' });
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Error desconocido.';
      setSubmit({ status: 'error', message });
    }
  };

  const isLoading = submit.status === 'loading';

  return (
    <form
      onSubmit={handleSubmit}
      className={styles.form}
      noValidate
      aria-label="Formulario de ingreso rápido"
    >
      <div className={styles.field}>
        <label htmlFor={typeId} className={styles.label}>
          Tipo
        </label>
        <select
          id={typeId}
          name="type"
          value={form.type}
          onChange={handleChange}
          className={styles.select}
          disabled={isLoading}
        >
          <option value="expense">💸 Gasto</option>
          <option value="income">💰 Ingreso</option>
        </select>
      </div>

      <div className={styles.field}>
        <label htmlFor={descId} className={styles.label}>
          Descripción
        </label>
        <input
          id={descId}
          name="description"
          type="text"
          placeholder="ej. Café, Gasolina, Uber…"
          value={form.description}
          onChange={handleChange}
          className={styles.input}
          disabled={isLoading}
          autoComplete="off"
          aria-describedby={submit.status === 'error' ? 'form-error' : undefined}
        />
      </div>

      <div className={styles.field}>
        <label htmlFor={amountId} className={styles.label}>
          Monto (USD)
        </label>
        <input
          id={amountId}
          name="amount"
          type="number"
          placeholder="0.00"
          min="0.01"
          step="0.01"
          value={form.amount}
          onChange={handleChange}
          className={styles.input}
          disabled={isLoading}
        />
      </div>

      {submit.status === 'error' && (
        <p id="form-error" role="alert" className={styles.errorMsg}>
          {submit.message}
        </p>
      )}
      {submit.status === 'success' && (
        <p role="status" className={styles.successMsg}>
          {submit.message}
        </p>
      )}

      <button
        type="submit"
        disabled={isLoading}
        className={styles.submitBtn}
        aria-busy={isLoading}
      >
        {isLoading ? (
          <span className={styles.spinner} aria-hidden="true" />
        ) : (
          '+ Registrar Gasto'
        )}
      </button>
    </form>
  );
}
