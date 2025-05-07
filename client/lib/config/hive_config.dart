import 'package:hive_flutter/hive_flutter.dart';
import '../models/user.dart';

class HiveConfig {
  static const String userBoxName = 'userBox';
  static const int userTypeId = 0;
  static const String userKey = 'current_user';

  static Future<void> init() async {
    await Hive.initFlutter();
    _registerAdapters();
    await _openBoxes();
  }

  static void _registerAdapters() {
    Hive.registerAdapter(UserAdapter());
  }

  static Future<void> _openBoxes() async {
    await Hive.openBox<User>(userBoxName);
  }

  static Box<User> getUserBox() {
    return Hive.box<User>(userBoxName);
  }
}
