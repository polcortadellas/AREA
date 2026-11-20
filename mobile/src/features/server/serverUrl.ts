/**
 * Parses the server address typed by the user into a normalised base URL
 * (`http://host:port[/path]`, no trailing slash).
 *
 * React Native's `URL` polyfill does not implement the getters we need
 * (hostname, port...), so the address is parsed by hand.
 */

export type ServerUrlError = 'required' | 'scheme' | 'host' | 'port';

export type ParseResult =
  { ok: true; url: string } | { ok: false; error: ServerUrlError };

/** Android emulator alias for the host machine's localhost. */
export const EMULATOR_HOST_URL = 'http://10.0.2.2:8080';

const SCHEME = /^([a-z][a-z0-9+.-]*):\/\//i;
// [IPv6] | hostname / IPv4, then optional :port and /path
const AUTHORITY =
  /^(\[[0-9a-f:.]+\]|[^:/?#[\]@\s]+)(?::([^/?#]*))?(\/[^?#]*)?$/i;
const IPV4 = /^\d{1,3}(\.\d{1,3}){3}$/;
const HOST_LABEL = /^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$/i;

function isValidHost(host: string): boolean {
  if (host.startsWith('[')) {
    return host.length > 2;
  }
  if (IPV4.test(host)) {
    return host.split('.').every(part => Number(part) <= 255);
  }
  // A hostname made only of digits and dots is a mistyped IPv4 address.
  if (/^[\d.]+$/.test(host)) {
    return false;
  }
  return (
    host.length <= 253 && host.split('.').every(label => HOST_LABEL.test(label))
  );
}

export function parseServerUrl(input: string): ParseResult {
  let value = input.trim();
  if (value === '') {
    return { ok: false, error: 'required' };
  }

  // "192.168.1.20:8080" is what people type: assume http.
  const scheme = SCHEME.exec(value);
  if (!scheme) {
    value = `http://${value}`;
  } else if (!['http', 'https'].includes(scheme[1].toLowerCase())) {
    return { ok: false, error: 'scheme' };
  }

  const [, rawScheme, rest] = /^([^:]+):\/\/(.*)$/.exec(value)!;
  const match = AUTHORITY.exec(rest);
  if (!match) {
    return { ok: false, error: 'host' };
  }
  const [, rawHost, port, path = ''] = match;
  const host = rawHost.toLowerCase();
  if (!isValidHost(host)) {
    return { ok: false, error: 'host' };
  }

  if (port !== undefined) {
    const n = Number(port);
    if (!/^\d{1,5}$/.test(port) || n < 1 || n > 65535) {
      return { ok: false, error: 'port' };
    }
  }

  const url = `${rawScheme.toLowerCase()}://${host}${port ? `:${Number(port)}` : ''}${path}`;
  return { ok: true, url: url.replace(/\/+$/, '') };
}
