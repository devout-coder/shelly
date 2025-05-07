import 'package:flutter_bloc/flutter_bloc.dart';
import '../../services/shell_service.dart';
import 'shell_state.dart';

class ShellCubit extends Cubit<ShellState> {
  final ShellService _shellService;

  ShellCubit(this._shellService) : super(ShellInitial());

  Future<void> initializeShell() async {
    emit(ShellLoading());
    try {
      await _shellService.createShell();
      final channel = _shellService.connectToShell();
      emit(ShellConnected(channel));
    } catch (e) {
      emit(ShellError(e.toString()));
    }
  }

  Future<void> deleteShell() async {
    try {
      await _shellService.deleteShell();
      emit(ShellInitial());
    } catch (e) {
      emit(ShellError(e.toString()));
    }
  }
}
