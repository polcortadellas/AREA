import AsyncStorage from '@react-native-async-storage/async-storage';
import {
  fireEvent,
  render,
  screen,
  waitFor,
} from '@testing-library/react-native';

import App from './App';
import { i18n } from './i18n';

const pong = () =>
  Promise.resolve({
    ok: true,
    status: 200,
    text: () => Promise.resolve('pong\n'),
  } as Response);

beforeEach(async () => {
  await AsyncStorage.clear();
  await i18n.changeLanguage('en');
  globalThis.fetch = jest.fn(pong) as jest.Mock;
});

describe('App', () => {
  it('asks for the server address on first launch, then opens Home', async () => {
    await render(<App />);

    const input = await screen.findByLabelText('Server address');
    expect(screen.getByRole('header', { name: 'Server' })).toBeOnTheScreen();

    await fireEvent.changeText(input, 'not a host!');
    await fireEvent.press(screen.getByRole('button', { name: 'Save' }));
    expect(
      await screen.findByText('Enter a valid host name or IP address.'),
    ).toBeOnTheScreen();

    await fireEvent.changeText(input, '192.168.1.20:8080');
    await fireEvent.press(screen.getByRole('button', { name: 'Save' }));

    expect(
      await screen.findByText('http://192.168.1.20:8080'),
    ).toBeOnTheScreen();
    await waitFor(() =>
      expect(screen.getByText('Connected to the server.')).toBeOnTheScreen(),
    );
    expect(globalThis.fetch).toHaveBeenCalledWith(
      'http://192.168.1.20:8080/ping',
      expect.anything(),
    );
    await expect(AsyncStorage.getItem('area.serverUrl')).resolves.toBe(
      'http://192.168.1.20:8080',
    );
  });

  it('opens Home directly when an address is saved and reports an unreachable server', async () => {
    await AsyncStorage.setItem('area.serverUrl', 'http://10.0.2.2:8080');
    globalThis.fetch = jest
      .fn()
      .mockRejectedValue(new TypeError('Network request failed'));

    await render(<App />);

    expect(await screen.findByText('http://10.0.2.2:8080')).toBeOnTheScreen();
    expect(
      await screen.findByText(
        'Could not reach the server. Check the address and that the server is running.',
      ),
    ).toBeOnTheScreen();
    expect(
      screen.getByRole('button', { name: 'Check again' }),
    ).toBeOnTheScreen();
  });

  it('switches to Spanish and remembers the choice', async () => {
    await render(<App />);

    const spanish = await screen.findByRole('radio', { name: 'Español' });
    expect(spanish).not.toBeChecked();
    await fireEvent.press(spanish);

    expect(
      await screen.findByLabelText('Dirección del servidor'),
    ).toBeOnTheScreen();
    expect(screen.getByRole('radio', { name: 'Español' })).toBeChecked();
    await expect(AsyncStorage.getItem('area.lang')).resolves.toBe('es');
  });
});
