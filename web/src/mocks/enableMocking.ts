/**
 * Starts the MSW service worker when running `npm run dev` with VITE_API_MOCKING=true.
 * The dynamic import keeps MSW and the mock data out of production bundles.
 */
export async function enableMocking() {
  if (!import.meta.env.DEV || import.meta.env.VITE_API_MOCKING !== 'true') {
    return;
  }

  const { worker } = await import('./browser');
  await worker.start({ onUnhandledFrame: 'bypass' });
}
