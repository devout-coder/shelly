import 'package:equatable/equatable.dart';
import 'package:web_socket_channel/web_socket_channel.dart';

abstract class ShellState extends Equatable {
  const ShellState();

  @override
  List<Object?> get props => [];
}

class ShellInitial extends ShellState {}

class ShellLoading extends ShellState {}

class ShellConnected extends ShellState {
  final WebSocketChannel channel;

  const ShellConnected(this.channel);

  @override
  List<Object?> get props => [channel];
}

class ShellError extends ShellState {
  final String message;

  const ShellError(this.message);

  @override
  List<Object?> get props => [message];
}
