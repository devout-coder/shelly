import 'dart:convert';
import '../models/user.dart';
import 'request_service.dart';

class AuthService {
  Future<User> signup(String email, String password) async {
    final response = await RequestService.request(
      endpoint: '/auth/signup',
      method: 'POST',
      body: {
        'email': email,
        'password': password,
      },
    );
    final responseData = jsonDecode(response.body);
    if (responseData['success'] == true) {
      return User.fromJson(responseData['user']);
    } else {
      throw Exception(responseData['message'] ?? 'Failed to sign up');
    }
  }

  Future<User> login(String email, String password) async {
    final response = await RequestService.request(
      endpoint: '/auth/login',
      method: 'POST',
      body: {
        'email': email,
        'password': password,
      },
    );
    final responseData = jsonDecode(response.body);
    if (responseData['success'] == true) {
      return User.fromJson(responseData['user']);
    } else {
      throw Exception(responseData['message'] ?? 'Failed to login');
    }
  }
}
