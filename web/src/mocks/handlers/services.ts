import { http, HttpResponse } from 'msw';

import type { LinkedService } from '@/shared/api/generated/model';

/**
 * Behave like the /services routes (M-14) in `npm run dev` and tests. Only the
 * services in `mockOAuthServices` have an OAuth app; "authorize" links the
 * service at once and returns the address the real callback redirects to.
 * Links are kept for the browser tab, since that redirect reloads the page.
 */
export const mockOAuthServices = ['github'];
export const mockScopes = 'repo read:user';

const storageKey = 'area.mock.services';

function readLinks(): LinkedService[] {
  try {
    return JSON.parse(
      window.sessionStorage.getItem(storageKey) ?? '[]',
    ) as LinkedService[];
  } catch {
    return [];
  }
}

export function setMockLinks(links: LinkedService[]) {
  try {
    window.sessionStorage.setItem(storageKey, JSON.stringify(links));
  } catch {
    // Storage unavailable: links simply do not persist.
  }
}

const notFound = () =>
  HttpResponse.json({ error: 'not_found' }, { status: 404 });

export const servicesHandlers = [
  http.get('*/services', () => HttpResponse.json(readLinks())),
  http.post('*/services/:service/authorize', ({ params }) => {
    const service = String(params.service);
    if (!mockOAuthServices.includes(service)) return notFound();
    setMockLinks([
      ...readLinks().filter(l => l.service !== service),
      { service, scopes: mockScopes, linked_at: new Date().toISOString() },
    ]);
    return HttpResponse.json({
      url: `${window.location.origin}/services?linked=${service}`,
    });
  }),
  http.delete('*/services/:service', ({ params }) => {
    const links = readLinks();
    const rest = links.filter(l => l.service !== params.service);
    if (rest.length === links.length) return notFound();
    setMockLinks(rest);
    return new HttpResponse(null, { status: 204 });
  }),
];
