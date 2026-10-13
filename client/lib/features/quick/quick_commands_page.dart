import 'package:flutter/material.dart';

import 'package:hitest/api/v2/quick_command_v2.dart';
import 'package:hitest/core/i18n/l10n_x.dart';
import 'package:hitest/core/network/api_client_manager.dart';
import 'package:hitest/core/theme/app_design_tokens.dart';
import 'package:hitest/shared/widgets/operations/module_empty_state_widget.dart';
import 'package:hitest/shared/widgets/operations/module_error_state_widget.dart';

/// 快捷命令页：列出并执行服务端预置命令（白名单来自 hitest/conf/commands.yaml）。
class QuickCommandsPage extends StatefulWidget {
  const QuickCommandsPage({super.key});

  @override
  State<QuickCommandsPage> createState() => _QuickCommandsPageState();
}

class _QuickCommandsPageState extends State<QuickCommandsPage> {
  List<Map<String, dynamic>>? _commands;
  String? _error;
  String? _runningName;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<QuickCommandV2Api> _ensureApi() async {
    final client = await ApiClientManager().getCurrentClient();
    return QuickCommandV2Api(client);
  }

  Future<void> _load() async {
    setState(() {
      _error = null;
    });
    try {
      final api = await _ensureApi();
      final commands = (await api.getPresetCommands()).data ?? const [];
      if (!mounted) return;
      setState(() {
        _commands = commands;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e.toString();
      });
    }
  }

  Future<void> _run(Map<String, dynamic> command) async {
    final name = command['name']?.toString() ?? '';
    if (name.isEmpty || _runningName != null) return;
    final l10n = context.l10n;
    final messenger = ScaffoldMessenger.of(context);

    setState(() {
      _runningName = name;
    });
    String? output;
    Object? runError;
    try {
      final api = await _ensureApi();
      output = (await api.runPresetCommand(name)).data ?? '';
    } catch (e) {
      runError = e;
    }
    if (!mounted) return;
    setState(() {
      _runningName = null;
    });

    await showModalBottomSheet<void>(
      context: context,
      showDragHandle: true,
      isScrollControlled: true,
      builder: (sheetContext) {
        final theme = Theme.of(sheetContext);
        return SafeArea(
          child: Padding(
            padding: EdgeInsets.fromLTRB(
              AppDesignTokens.spacingLg,
              AppDesignTokens.spacingSm,
              AppDesignTokens.spacingLg,
              AppDesignTokens.spacingLg,
            ),
            child: ConstrainedBox(
              constraints: BoxConstraints(
                maxHeight: MediaQuery.of(sheetContext).size.height * 0.6,
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(name, style: theme.textTheme.titleMedium),
                  const SizedBox(height: AppDesignTokens.spacingMd),
                  Expanded(
                    child: Container(
                      width: double.infinity,
                      padding: const EdgeInsets.all(AppDesignTokens.spacingMd),
                      decoration: BoxDecoration(
                        color: theme.colorScheme.surfaceContainerLow,
                        borderRadius:
                            BorderRadius.circular(AppDesignTokens.radiusMd),
                      ),
                      child: SingleChildScrollView(
                        child: SelectableText(
                          runError != null
                              ? '${l10n.commonLoadFailedTitle}: $runError'
                              : (output?.isEmpty ?? true)
                                  ? l10n.quickCommandsNoOutput
                                  : output!,
                          style: theme.textTheme.bodySmall?.copyWith(
                            fontFamily: 'monospace',
                          ),
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        );
      },
    );
    messenger.hideCurrentSnackBar();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = context.l10n;
    return Scaffold(
      appBar: AppBar(title: Text(l10n.quickCommandsTitle)),
      body: _buildBody(context, l10n),
    );
  }

  Widget _buildBody(BuildContext context, dynamic l10n) {
    if (_error != null) {
      return ModuleErrorStateWidget(
        message: _error!,
        onRetry: _load,
      );
    }
    final commands = _commands;
    if (commands == null) {
      return const Center(child: CircularProgressIndicator());
    }
    if (commands.isEmpty) {
      return ModuleEmptyStateWidget(
        title: l10n.quickCommandsTitle,
        description: l10n.quickCommandsEmptyDesc,
        icon: Icons.terminal_outlined,
      );
    }
    return RefreshIndicator(
      onRefresh: _load,
      child: ListView.separated(
        padding: AppDesignTokens.pagePadding,
        itemCount: commands.length,
        separatorBuilder: (_, __) =>
            const SizedBox(height: AppDesignTokens.spacingSm),
        itemBuilder: (context, index) {
          final command = commands[index];
          final name = command['name']?.toString() ?? '';
          final isRunning = _runningName == name;
          return Card(
            margin: EdgeInsets.zero,
            child: ListTile(
              leading: Icon(
                Icons.terminal_outlined,
                color: Theme.of(context).colorScheme.primary,
              ),
              title: Text(name),
              subtitle: Text(
                command['command']?.toString() ?? '',
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: Theme.of(context).textTheme.bodySmall,
              ),
              trailing: isRunning
                  ? const SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : FilledButton.tonal(
                      onPressed: () => _run(command),
                      child: Text(l10n.commonConfirm),
                    ),
              onTap: isRunning ? null : () => _run(command),
            ),
          );
        },
      ),
    );
  }
}
