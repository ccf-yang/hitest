import 'package:flutter/material.dart';
import 'package:hitest/core/i18n/l10n_x.dart';
import 'package:hitest/core/theme/app_design_tokens.dart';
import 'package:hitest/shared/widgets/operations/module_empty_state_widget.dart';
import 'package:hitest/shared/widgets/operations/module_error_state_widget.dart';
import 'package:hitest/features/quick/quick_commands_page.dart';
import 'package:hitest/features/settings/monitor_settings_page.dart';
import 'package:hitest/shared/widgets/app_card.dart';
import 'package:provider/provider.dart';
import '../../data/models/monitoring_runtime_models.dart';
import '../../data/repositories/monitor_repository.dart';
import 'monitoring_provider.dart';
import 'widgets/monitor_chart.dart';

class MonitoringPage extends StatefulWidget {
  const MonitoringPage({super.key});

  @override
  State<MonitoringPage> createState() => _MonitoringPageState();
}

class _MonitoringPageState extends State<MonitoringPage> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      final provider = context.read<MonitoringProvider>();
      provider.load();
      // 默认启用自动刷新
      provider.toggleAutoRefresh(true);
    });
  }

  @override
  Widget build(BuildContext context) {
    final l10n = context.l10n;
    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.serverModuleMonitoring),
        actions: [
          Consumer<MonitoringProvider>(
            builder: (context, provider, _) => IconButton(
              icon: const Icon(Icons.refresh),
              onPressed: provider.data.isLoading ? null : provider.refresh,
              tooltip: l10n.commonRefresh,
            ),
          ),
          IconButton(
            icon: const Icon(Icons.tune),
            tooltip: l10n.monitorSettings,
            onPressed: () => Navigator.push(
              context,
              MaterialPageRoute(
                builder: (_) => const MonitorSettingsPage(),
              ),
            ),
          ),
          IconButton(
            icon: const Icon(Icons.terminal_outlined),
            tooltip: l10n.quickCommandsTitle,
            onPressed: () => Navigator.push(
              context,
              MaterialPageRoute(
                builder: (_) => const QuickCommandsPage(),
              ),
            ),
          ),
        ],
      ),
      body: Padding(
        padding: AppDesignTokens.pagePadding,
        child: Consumer<MonitoringProvider>(
          builder: (context, provider, _) {
            return _buildBody(context, provider.data, provider.refresh);
          },
        ),
      ),
    );
  }

  Widget _buildBody(
    BuildContext context,
    MonitoringData data,
    Future<void> Function() onRefresh,
  ) {
    final l10n = context.l10n;
    final hasData = data.currentMetrics != null ||
        data.cpuTimeSeries != null ||
        data.memoryTimeSeries != null;

    if (data.error != null && !hasData) {
      return ModuleErrorStateWidget(
        message: data.error!,
        onRetry: onRefresh,
      );
    }

    if (data.isLoading && !hasData) {
      return const Center(child: CircularProgressIndicator());
    }

    return RefreshIndicator(
      onRefresh: onRefresh,
      child: ListView(
        children: [
          _buildCurrentMetrics(context, data.currentMetrics),
          const SizedBox(height: AppDesignTokens.spacingMd),
          _buildTimeSeriesCard(
            context,
            l10n.serverCpuLabel,
            data.cpuTimeSeries,
            '%',
          ),
          const SizedBox(height: AppDesignTokens.spacingSm),
          _buildTimeSeriesCard(
            context,
            l10n.serverMemoryLabel,
            data.memoryTimeSeries,
            '%',
          ),
          const SizedBox(height: AppDesignTokens.spacingSm),
          _buildTimeSeriesCard(
            context,
            l10n.serverLoadLabel,
            data.loadTimeSeries,
            '',
          ),
          const SizedBox(height: AppDesignTokens.spacingSm),
          _buildTimeSeriesCard(
            context,
            '${l10n.serverDiskLabel} IO',
            data.ioTimeSeries,
            'KB/s',
            selector: _MetricSelector(
              value: data.selectedIO,
              options: data.ioOptions,
              onChanged: (value) =>
                  context.read<MonitoringProvider>().selectIOOption(value),
            ),
          ),
          const SizedBox(height: AppDesignTokens.spacingSm),
          _buildTimeSeriesCard(
            context,
            l10n.monitorNetworkLabel,
            data.networkTimeSeries,
            'KB/s',
            selector: _MetricSelector(
              value: data.selectedNetwork,
              options: data.networkOptions,
              onChanged: (value) =>
                  context.read<MonitoringProvider>().selectNetworkOption(value),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCurrentMetrics(
      BuildContext context, MonitorMetricsSnapshot? metrics) {
    final l10n = context.l10n;
    if (metrics == null) return const SizedBox.shrink();

    return AppCard(
      title: l10n.monitorMetricCurrent,
      child: GridView.count(
        shrinkWrap: true,
        physics: const NeverScrollableScrollPhysics(),
        crossAxisCount: 2,
        mainAxisSpacing: AppDesignTokens.spacingSm,
        crossAxisSpacing: AppDesignTokens.spacingSm,
        childAspectRatio: 3.5,
        children: [
          _MetricChip(
            label: l10n.serverCpuLabel,
            value: metrics.cpuPercent != null
                ? '${metrics.cpuPercent!.toStringAsFixed(1)}%'
                : '--',
            icon: Icons.memory_outlined,
          ),
          _MetricChip(
            label: l10n.serverMemoryLabel,
            value: metrics.memoryPercent != null
                ? '${metrics.memoryPercent!.toStringAsFixed(1)}%'
                : '--',
            icon: Icons.storage_outlined,
          ),
          _MetricChip(
            label: l10n.serverDiskLabel,
            value: metrics.diskPercent != null
                ? '${metrics.diskPercent!.toStringAsFixed(1)}%'
                : '--',
            icon: Icons.folder_outlined,
          ),
          _MetricChip(
            label: l10n.serverLoadLabel,
            value: metrics.load1 != null
                ? metrics.load1!.toStringAsFixed(2)
                : '--',
            icon: Icons.speed_outlined,
          ),
        ],
      ),
    );
  }

  Widget _buildTimeSeriesCard(BuildContext context, String title,
      MonitorTimeSeries? timeSeries, String unit,
      {Widget? selector}) {
    return _ExpandableChartCard(
      title: title,
      timeSeries: timeSeries,
      unit: unit,
      selector: selector,
    );
  }
}

class _MetricChip extends StatelessWidget {
  const _MetricChip({
    required this.label,
    required this.value,
    required this.icon,
  });

  final String label;
  final String value;
  final IconData icon;

  @override
  Widget build(BuildContext context) {
    return Chip(
      avatar: Icon(icon, size: 18),
      label: Text('$label: $value'),
    );
  }
}

class _StatItem extends StatelessWidget {
  const _StatItem({
    required this.label,
    required this.value,
  });

  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;

    return Container(
      padding: const EdgeInsets.symmetric(
        horizontal: AppDesignTokens.spacingSm,
        vertical: AppDesignTokens.spacingSm,
      ),
      decoration: BoxDecoration(
        color: colorScheme.surfaceContainerHigh,
        borderRadius: BorderRadius.circular(AppDesignTokens.radiusMd),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            label,
            style: theme.textTheme.labelMedium?.copyWith(
              color: colorScheme.onSurfaceVariant,
            ),
          ),
          const SizedBox(height: AppDesignTokens.spacingXs),
          Text(
            value,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            style: theme.textTheme.titleMedium?.copyWith(
              fontWeight: FontWeight.w700,
            ),
          ),
        ],
      ),
    );
  }
}

class _ExpandableChartCard extends StatefulWidget {
  final String title;
  final MonitorTimeSeries? timeSeries;
  final String unit;
  final Widget? selector;

  const _ExpandableChartCard({
    required this.title,
    required this.timeSeries,
    required this.unit,
    this.selector,
  });

  @override
  State<_ExpandableChartCard> createState() => _ExpandableChartCardState();
}

class _ExpandableChartCardState extends State<_ExpandableChartCard> {
  bool _isExpanded = true;

  @override
  Widget build(BuildContext context) {
    final l10n = context.l10n;
    final timeSeries = widget.timeSeries;
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;
    final currentValue = timeSeries != null && timeSeries.data.isNotEmpty
        ? '${timeSeries.data.last.value.toStringAsFixed(1)}${widget.unit}'
        : null;

    return Card(
      clipBehavior: Clip.antiAlias,
      elevation: 0,
      color: colorScheme.surfaceContainerLow,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(AppDesignTokens.radiusLg),
        side: BorderSide(color: colorScheme.outlineVariant),
      ),
      child: Column(
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(
              AppDesignTokens.spacingMd,
              AppDesignTokens.spacingMd,
              AppDesignTokens.spacingMd,
              AppDesignTokens.spacingSm,
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            widget.title,
                            style: theme.textTheme.titleMedium?.copyWith(
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                          const SizedBox(height: AppDesignTokens.spacingXs),
                          Text(
                            currentValue ?? l10n.commonEmpty,
                            style: theme.textTheme.headlineSmall?.copyWith(
                              fontWeight: FontWeight.w700,
                              color: currentValue != null
                                  ? colorScheme.onSurface
                                  : colorScheme.onSurfaceVariant,
                            ),
                          ),
                        ],
                      ),
                    ),
                    IconButton.filledTonal(
                      onPressed: () {
                        setState(() {
                          _isExpanded = !_isExpanded;
                        });
                      },
                      icon: AnimatedRotation(
                        turns: _isExpanded ? 0.5 : 0,
                        duration: const Duration(milliseconds: 200),
                        child: const Icon(Icons.keyboard_arrow_down_rounded),
                      ),
                    ),
                  ],
                ),
                if (widget.selector != null) ...[
                  const SizedBox(height: AppDesignTokens.spacingSm),
                  widget.selector!,
                ],
              ],
            ),
          ),
          AnimatedCrossFade(
            firstChild: const SizedBox.shrink(),
            secondChild: timeSeries == null || timeSeries.data.isEmpty
                ? ModuleEmptyStateWidget(
                    title: l10n.commonEmpty,
                    description: widget.title.contains('IO')
                        ? '请在设置中配置磁盘IO设备（如 /dev/vda2）'
                        : widget.title.contains('网络')
                            ? '请在设置中配置网络接口（如 eth0）'
                            : l10n.commonComingSoon,
                  )
                : Padding(
                    padding: const EdgeInsets.fromLTRB(
                      AppDesignTokens.spacingMd,
                      AppDesignTokens.spacingSm,
                      AppDesignTokens.spacingMd,
                      AppDesignTokens.spacingMd,
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        _buildStatsRow(context, timeSeries, widget.unit),
                        const SizedBox(height: AppDesignTokens.spacingMd),
                        _buildSimpleChart(context, timeSeries, widget.unit),
                      ],
                    ),
                  ),
            crossFadeState: _isExpanded
                ? CrossFadeState.showSecond
                : CrossFadeState.showFirst,
            duration: const Duration(milliseconds: 200),
          ),
        ],
      ),
    );
  }

  Widget _buildStatsRow(
    BuildContext context,
    MonitorTimeSeries timeSeries,
    String unit,
  ) {
    final l10n = context.l10n;

    return Row(
      children: [
        Expanded(
          child: _StatItem(
            label: l10n.monitorMetricMin,
            value: timeSeries.min != null
                ? '${timeSeries.min!.toStringAsFixed(1)}$unit'
                : '--',
          ),
        ),
        const SizedBox(width: AppDesignTokens.spacingSm),
        Expanded(
          child: _StatItem(
            label: l10n.monitorMetricAvg,
            value: timeSeries.avg != null
                ? '${timeSeries.avg!.toStringAsFixed(1)}$unit'
                : '--',
          ),
        ),
        const SizedBox(width: AppDesignTokens.spacingSm),
        Expanded(
          child: _StatItem(
            label: l10n.monitorMetricMax,
            value: timeSeries.max != null
                ? '${timeSeries.max!.toStringAsFixed(1)}$unit'
                : '--',
          ),
        ),
      ],
    );
  }

  Widget _buildSimpleChart(
    BuildContext context,
    MonitorTimeSeries timeSeries,
    String unit,
  ) {
    if (timeSeries.data.isEmpty) return const SizedBox.shrink();

    return Container(
      padding: const EdgeInsets.fromLTRB(
        AppDesignTokens.spacingSm,
        AppDesignTokens.spacingSm,
        AppDesignTokens.spacingSm,
        AppDesignTokens.spacingXs,
      ),
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surface,
        borderRadius: BorderRadius.circular(AppDesignTokens.radiusMd),
      ),
      child: SizedBox(
        height: 220,
        child: MonitorChart(
          data: timeSeries.data,
          unit: unit,
          label: timeSeries.name,
          color: Theme.of(context).colorScheme.primary,
        ),
      ),
    );
  }
}


class _MetricSelector extends StatelessWidget {
  const _MetricSelector({
    required this.value,
    required this.options,
    required this.onChanged,
  });

  final String value;
  final List<String> options;
  final ValueChanged<String> onChanged;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;
    final items = options.isEmpty ? const ['all'] : options;
    final currentValue = items.contains(value) ? value : items.first;

    return Container(
      constraints: const BoxConstraints(maxWidth: 220),
      padding:
          const EdgeInsets.symmetric(horizontal: AppDesignTokens.spacingSm),
      decoration: BoxDecoration(
        color: colorScheme.surfaceContainerHighest,
        border: Border.all(color: colorScheme.outlineVariant),
        borderRadius: BorderRadius.circular(AppDesignTokens.radiusMd),
      ),
      child: DropdownButtonHideUnderline(
        child: DropdownButton<String>(
          value: currentValue,
          isExpanded: true,
          borderRadius: BorderRadius.circular(AppDesignTokens.radiusMd),
          icon: const Icon(Icons.arrow_drop_down_rounded),
          style: theme.textTheme.bodyMedium?.copyWith(
            color: colorScheme.onSurface,
          ),
          items: items
              .map(
                (item) => DropdownMenuItem<String>(
                  value: item,
                  child: Text(item, overflow: TextOverflow.ellipsis),
                ),
              )
              .toList(),
          onChanged: (next) {
            if (next != null) {
              onChanged(next);
            }
          },
        ),
      ),
    );
  }
}
