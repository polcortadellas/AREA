import 'package:flutter_test/flutter_test.dart';
import 'package:area_frontend/main.dart';

void main() {
  testWidgets('Renders PingScreen', (WidgetTester tester) async {
    await tester.pumpWidget(const AreaApp(autoPing: false));
    expect(find.text('Ping /ping'), findsOneWidget);
  });
}
