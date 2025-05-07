import 'dart:convert';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'request_service.dart';

class ShellService {
  Future<void> createShell() async {
    final response = await RequestService.request(
      endpoint: '/shell',
      method: 'POST',
    );
    final responseData = jsonDecode(response.body);
    if (responseData['success'] != true) {
      throw Exception(responseData['message'] ?? 'Failed to create shell');
    }
  }

  Future<void> deleteShell() async {
    final response = await RequestService.request(
      endpoint: '/shell',
      method: 'DELETE',
    );
    final responseData = jsonDecode(response.body);
    if (responseData['success'] != true) {
      throw Exception(responseData['message'] ?? 'Failed to delete shell');
    }
  }

  WebSocketChannel connectToShell() {
    return RequestService.connectWebSocket('/shell/ws');
  }
}
