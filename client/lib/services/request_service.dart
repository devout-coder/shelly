import 'dart:convert';
import 'dart:io';
import 'package:http/http.dart' as http;
import 'package:flutter_dotenv/flutter_dotenv.dart';
import '../config/hive_config.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:web_socket_channel/io.dart';

class RequestService {
  static final String wsBaseUrl = "ws://${dotenv.env['IP_PORT']}";
  static final String httpBaseUrl = 'http://${dotenv.env['IP_PORT']}';

  static Future<http.Response> request({
    required String endpoint,
    required String method,
    Map<String, dynamic>? body,
  }) async {
    final userBox = HiveConfig.getUserBox();
    final user = userBox.get(HiveConfig.userKey);

    final headers = {
      'Content-Type': 'application/json',
      if (user != null) 'Authorization': 'Bearer ${user.token}',
    };

    final uri = Uri.parse('$httpBaseUrl$endpoint');

    switch (method.toUpperCase()) {
      case 'GET':
        return await http.get(uri, headers: headers);
      case 'POST':
        return await http.post(
          uri,
          headers: headers,
          body: body != null ? jsonEncode(body) : null,
        );
      case 'PUT':
        return await http.put(
          uri,
          headers: headers,
          body: body != null ? jsonEncode(body) : null,
        );
      case 'DELETE':
        return await http.delete(
          uri,
          headers: headers,
          body: body != null ? jsonEncode(body) : null,
        );
      default:
        throw Exception('Unsupported HTTP method: $method');
    }
  }

  static WebSocketChannel connectWebSocket(String endpoint) {
    final userBox = HiveConfig.getUserBox();
    final user = userBox.get(HiveConfig.userKey);

    final token = user != null ? 'Bearer ${user.token}' : '';
    final wsUrl = Uri.parse('$wsBaseUrl$endpoint');

    final headers = {
      if (user != null) 'Authorization': token,
    };

    final socket = WebSocket.connect(
      wsUrl.toString(),
      headers: headers,
    );

    return IOWebSocketChannel(socket);
  }
}
