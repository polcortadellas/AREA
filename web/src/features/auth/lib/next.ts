/**
 * Where to go after logging in. Only paths inside the app are accepted, so a
 * crafted `?next=//evil.example` link cannot send the user to another site.
 */
export function safeNext(next: string | null): string {
  if (!next || !next.startsWith('/') || /^\/[/\\]/.test(next)) {
    return '/';
  }
  return next;
}
