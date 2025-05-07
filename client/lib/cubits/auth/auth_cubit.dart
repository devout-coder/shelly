import 'package:flutter_bloc/flutter_bloc.dart';
import '../../services/auth_service.dart';
import '../../services/shell_service.dart';
import 'auth_state.dart';
import '../../config/hive_config.dart';

class AuthCubit extends Cubit<AuthState> {
  final AuthService _authService;
  final ShellService _shellService;

  AuthCubit(this._authService)
      : _shellService = ShellService(),
        super(AuthInitial());

  Future<void> init() async {
    final userBox = HiveConfig.getUserBox();
    final user = userBox.get(HiveConfig.userKey);
    if (user != null) {
      emit(AuthAuthenticated(user));
    } else {
      emit(AuthUnauthenticated());
    }
  }

  Future<void> signup(String email, String password) async {
    emit(AuthLoading());
    try {
      final user = await _authService.signup(email, password);
      await HiveConfig.getUserBox().put(HiveConfig.userKey, user);
      emit(AuthAuthenticated(user));
    } catch (e) {
      emit(AuthError(e.toString()));
    }
  }

  Future<void> login(String email, String password) async {
    emit(AuthLoading());
    try {
      final user = await _authService.login(email, password);
      await HiveConfig.getUserBox().put(HiveConfig.userKey, user);
      emit(AuthAuthenticated(user));
    } catch (e) {
      emit(AuthError(e.toString()));
    }
  }

  Future<void> logout() async {
    try {
      await HiveConfig.getUserBox().delete(HiveConfig.userKey);
      emit(AuthUnauthenticated());
    } catch (e) {
      emit(AuthError(e.toString()));
    }
  }
}
