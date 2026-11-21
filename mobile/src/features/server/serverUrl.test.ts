import { parseServerUrl } from './serverUrl';

describe('parseServerUrl', () => {
  it.each([
    ['http://192.168.1.20:8080', 'http://192.168.1.20:8080'],
    ['  http://192.168.1.20:8080/  ', 'http://192.168.1.20:8080'],
    ['192.168.1.20:8080', 'http://192.168.1.20:8080'],
    ['10.0.2.2:8080', 'http://10.0.2.2:8080'],
    ['HTTPS://Area.Example.com', 'https://area.example.com'],
    ['https://area.example.com/api/', 'https://area.example.com/api'],
    ['localhost:08080', 'http://localhost:8080'],
    ['http://[::1]:8080', 'http://[::1]:8080'],
  ])('accepts %j as %j', (input, expected) => {
    expect(parseServerUrl(input)).toEqual({ ok: true, url: expected });
  });

  it.each([
    ['', 'required'],
    ['   ', 'required'],
    ['ftp://192.168.1.20', 'scheme'],
    ['http://', 'host'],
    ['http://256.1.1.1:8080', 'host'],
    ['http://192.168.1:8080', 'host'],
    ['http://my server:8080', 'host'],
    ['http://-bad-.com', 'host'],
    ['http://user@host:8080', 'host'],
    ['http://192.168.1.20:0', 'port'],
    ['http://192.168.1.20:65536', 'port'],
    ['http://192.168.1.20:abc', 'port'],
    ['http://192.168.1.20:', 'port'],
  ])('rejects %j with %j', (input, error) => {
    expect(parseServerUrl(input)).toEqual({ ok: false, error });
  });
});
