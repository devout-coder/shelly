import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:xterm/xterm.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:shelly/cubits/shell/shell_cubit.dart';
import 'package:shelly/cubits/shell/shell_state.dart';

class ShellScreen extends StatefulWidget {
  const ShellScreen({super.key});

  @override
  State<ShellScreen> createState() => _ShellScreenState();
}

class _ShellScreenState extends State<ShellScreen> {
  final _terminal = Terminal();
  final _terminalController = TerminalController();
  String _inputBuffer = '';
  StreamSubscription? _streamSubscription;
  WebSocketChannel? _currentChannel;

  @override
  void initState() {
    super.initState();
    _initializeShell();
  }

  void _initializeShell() {
    context.read<ShellCubit>().initializeShell();
  }

  @override
  void dispose() {
    _streamSubscription?.cancel();
    _terminalController.dispose();
    super.dispose();
  }

  void _handleBackspace() {
    if (_inputBuffer.isNotEmpty) {
      _inputBuffer = _inputBuffer.substring(0, _inputBuffer.length - 1);
      // Move cursor back one position and delete character
      _terminal.write('\x08 \x08');
    }
  }

  void _setupTerminal(WebSocketChannel channel) {
    if (_currentChannel == channel) return; // Skip if already set up
    _currentChannel = channel;

    // Cancel any existing subscription
    _streamSubscription?.cancel();

    // Set up new stream subscription
    _streamSubscription = channel.stream.listen((data) {
      _terminal.write(data);
    });

    _terminal.onOutput = (data) {
      // Handle backspace and delete keys
      if (data == '\x08' || data == '\x7f') {
        _handleBackspace();
        return;
      }

      // Handle enter key
      if (data == '\r' || data == '\n') {
        channel.sink.add(_inputBuffer + '\n'); // Send buffered input to server
        _inputBuffer = '';
        _terminal.write(data); // Show newline in terminal
        return;
      }

      // Buffer other characters
      _inputBuffer += data;
      _terminal.write(data); // Only update terminal display
    };
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Shell'),
        actions: [
          BlocBuilder<ShellCubit, ShellState>(
            builder: (context, state) {
              if (state is ShellConnected) {
                return IconButton(
                  icon: const Icon(Icons.delete),
                  onPressed: () async {
                    await context.read<ShellCubit>().deleteShell();
                    if (context.mounted) {
                      Navigator.pop(context);
                    }
                  },
                );
              }
              return const SizedBox.shrink();
            },
          ),
        ],
      ),
      body: BlocListener<ShellCubit, ShellState>(
        listener: (context, state) {
          if (state is ShellConnected) {
            _setupTerminal(state.channel);
          }
        },
        child: BlocBuilder<ShellCubit, ShellState>(
          builder: (context, state) {
            if (state is ShellLoading) {
              return const Center(
                child: CircularProgressIndicator(),
              );
            }
            if (state is ShellError) {
              return Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text(
                      'Error: ${state.message}',
                      style: const TextStyle(color: Colors.red),
                    ),
                    const SizedBox(height: 16),
                    ElevatedButton(
                      onPressed: _initializeShell,
                      child: const Text('Retry'),
                    ),
                  ],
                ),
              );
            }

            if (state is ShellConnected) {
              return TerminalView(_terminal);
            }

            return const Center(
              child: Text('No shell connection'),
            );
          },
        ),
      ),
    );
  }
}
