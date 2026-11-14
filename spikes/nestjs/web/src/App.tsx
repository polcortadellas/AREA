import { useEffect, useState, type FormEvent } from 'react';
import './App.css';

const API = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';

type Area = { id: number; name: string; action: string; reaction: string; lastTriggeredAt: string | null };

export default function App() {
  const [status, setStatus] = useState('checking…');
  const [areas, setAreas] = useState<Area[]>([]);
  const [error, setError] = useState<string | null>(null);

  const load = () =>
    fetch(`${API}/areas`).then((r) => r.json()).then(setAreas).catch(() => setError('cannot reach server'));

  useEffect(() => {
    fetch(`${API}/ping`)
      .then((r) => r.json())
      .then((d: { status: string }) => setStatus(d.status))
      .catch(() => setStatus('down'));
    load();
    const t = setInterval(load, 3000);
    return () => clearInterval(t);
  }, []);

  async function onSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const form = new FormData(e.currentTarget);
    const res = await fetch(`${API}/areas`, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify(Object.fromEntries(form)),
    });
    const body = await res.json();
    setError(res.ok ? null : [].concat(body.message).join(', '));
    if (res.ok) {
      e.currentTarget?.reset();
      load();
    }
  }

  return (
    <main>
      <h1>AREA — NestJS spike</h1>
      <p>Server: <strong>{status}</strong></p>

      <form onSubmit={onSubmit}>
        <input name="name" placeholder="name" />
        <input name="action" placeholder="action (e.g. github.new_issue)" />
        <input name="reaction" placeholder="reaction (e.g. discord.send_message)" />
        <button type="submit">Create AREA</button>
      </form>
      {error && <p style={{ color: 'crimson' }}>{error}</p>}

      <ul>
        {areas.map((a) => (
          <li key={a.id}>
            #{a.id} <b>{a.name}</b>: {a.action} → {a.reaction}{' '}
            <small>(last run: {a.lastTriggeredAt ? new Date(a.lastTriggeredAt).toLocaleTimeString() : 'never'})</small>
          </li>
        ))}
      </ul>
    </main>
  );
}
