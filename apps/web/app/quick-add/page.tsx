import { QuickAddForm } from '@/features/quick-add/components/QuickAddForm';

export const metadata = {
  title: 'Ingreso Rápido — Clan Budget',
  description: 'Registra gastos hormiga al instante con Clan Budget.',
};

export default function QuickAddPage() {
  return (
    <main
      style={{
        minHeight: '100dvh',
        background: '#1a1b26',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '2rem 1rem',
        fontFamily: "'Inter', system-ui, sans-serif",
        color: '#c0caf5',
      }}
    >
      <header style={{ textAlign: 'center', marginBottom: '2.5rem' }}>
        <p style={{ fontSize: '2.5rem', margin: 0 }}>⚡</p>
        <h1
          style={{
            margin: '0.5rem 0 0.25rem',
            fontSize: '1.75rem',
            fontWeight: 700,
            color: '#c0caf5',
          }}
        >
          Ingreso Rápido
        </h1>
        <p style={{ margin: 0, color: '#565f89', fontSize: '0.925rem' }}>
          Registra un gasto hormiga en segundos
        </p>
      </header>

      <QuickAddForm />
    </main>
  );
}
