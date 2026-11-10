import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;

const String apiBaseUrl = String.fromEnvironment('API_URL');

void main() {
  runApp(const AreaApp());
}

class AreaApp extends StatelessWidget {
  const AreaApp({super.key, this.autoPing = true});

  final bool autoPing;

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'AREA',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.indigo),
        useMaterial3: true,
      ),
      home: PingScreen(autoPing: autoPing),
    );
  }
}

class PingScreen extends StatefulWidget {
  const PingScreen({super.key, this.autoPing = true});

  final bool autoPing;

  @override
  State<PingScreen> createState() => _PingScreenState();
}

class _PingScreenState extends State<PingScreen> {
  String _rawBody = '';
  String? _statusValue;
  String? _error;
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    if (widget.autoPing) {
      _pingServer();
    }
  }

  Future<void> _pingServer() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final res = await http
          .get(Uri.parse('$apiBaseUrl/ping'))
          .timeout(const Duration(seconds: 5));
      String? parsedStatus;
      try {
        final decoded = jsonDecode(res.body);
        if (decoded is Map<String, dynamic>) {
          parsedStatus = decoded['status']?.toString();
        }
      } catch (_) {}

      if (!mounted) return;
      setState(() {
        _rawBody = res.body;
        _statusValue = parsedStatus;
        if (res.statusCode != 200) {
          _error = 'HTTP ${res.statusCode}';
        }
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e.toString();
        _rawBody = '';
        _statusValue = null;
      });
    } finally {
      if (mounted) {
        setState(() {
          _loading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final bool isConnected = _error == null && _statusValue == 'ok';

    return Scaffold(
      appBar: AppBar(
        title: const Text('AREA — Flutter + Hono/Bun + PostgreSQL'),
        centerTitle: true,
      ),
      body: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 460),
          child: Card(
            margin: const EdgeInsets.all(24),
            elevation: 2,
            child: Padding(
              padding: const EdgeInsets.all(24),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(
                    _loading
                        ? Icons.sync
                        : (isConnected ? Icons.check_circle : Icons.error_outline),
                    size: 56,
                    color: _loading
                        ? Colors.blueGrey
                        : (isConnected ? Colors.green : Colors.red),
                  ),
                  const SizedBox(height: 16),
                  Text(
                    _loading
                        ? 'Connecting to backend...'
                        : (isConnected
                            ? 'Backend Connected'
                            : 'Backend Unreachable'),
                    style: Theme.of(context).textTheme.headlineSmall,
                  ),
                  const SizedBox(height: 8),
                  Text(
                    'Endpoint: ${apiBaseUrl.isEmpty ? "/ping" : "$apiBaseUrl/ping"}',
                    style: Theme.of(context).textTheme.bodySmall,
                  ),
                  const SizedBox(height: 16),
                  Container(
                    width: double.infinity,
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: Colors.grey.shade100,
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(color: Colors.grey.shade300),
                    ),
                    child: SelectableText(
                      _error != null
                          ? 'Error: $_error'
                          : (_rawBody.isEmpty ? 'No response yet' : _rawBody),
                      textAlign: TextAlign.center,
                      style: const TextStyle(fontFamily: 'monospace'),
                    ),
                  ),
                  const SizedBox(height: 20),
                  FilledButton.icon(
                    onPressed: _loading ? null : _pingServer,
                    icon: const Icon(Icons.refresh),
                    label: const Text('Ping /ping'),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
