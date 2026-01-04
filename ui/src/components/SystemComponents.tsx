import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'

interface ComponentStatus {
  name: string
  id: string
  status: 'running' | 'degraded' | 'error' | 'unknown'
  description: string
  metrics: {
    label: string
    value: string | number
  }[]
  recentActivity: {
    time: string
    message: string
    type: 'info' | 'success' | 'warning' | 'error'
  }[]
  dependencies: string[]
  config: {
    key: string
    value: string
  }[]
}

const componentData: ComponentStatus[] = [
  {
    name: 'Falco Sensor',
    id: 'falco',
    status: 'running',
    description: 'DaemonSet collecting runtime events from container syscalls using eBPF',
    metrics: [
      { label: 'Events/sec', value: '~150' },
      { label: 'Drop Rate', value: '0.01%' },
      { label: 'CPU Usage', value: '2.3%' },
      { label: 'Memory', value: '256MB' },
    ],
    recentActivity: [
      { time: '2s ago', message: 'process_exec event captured', type: 'info' },
      { time: '5s ago', message: 'file_open event captured', type: 'info' },
      { time: '10s ago', message: 'Heartbeat sent', type: 'success' },
    ],
    dependencies: ['containerd', 'kernel'],
    config: [
      { key: 'Output', value: 'HTTP to Ingest API' },
      { key: 'Buffer Size', value: '8MB' },
      { key: 'Rules', value: 'podwatch_rules.yaml' },
    ],
  },
  {
    name: 'Ingest API',
    id: 'ingest',
    status: 'running',
    description: 'Receives events from Falco, validates schema, publishes to NATS, stores raw data in S3',
    metrics: [
      { label: 'Requests/sec', value: '~150' },
      { label: 'Latency p95', value: '12ms' },
      { label: 'Error Rate', value: '0%' },
      { label: 'Queue Depth', value: '0' },
    ],
    recentActivity: [
      { time: '1s ago', message: 'Event published to NATS', type: 'success' },
      { time: '3s ago', message: 'Batch written to S3', type: 'success' },
      { time: '8s ago', message: 'Schema validation passed', type: 'info' },
    ],
    dependencies: ['NATS', 'MinIO/S3'],
    config: [
      { key: 'Port', value: '8080' },
      { key: 'NATS Subject', value: 'events.raw.*' },
      { key: 'S3 Bucket', value: 'podwatch-raw' },
    ],
  },
  {
    name: 'Enrichment',
    id: 'enrich',
    status: 'running',
    description: 'Attaches Kubernetes metadata (pod, namespace, labels) to raw events',
    metrics: [
      { label: 'Events/sec', value: '~150' },
      { label: 'Cache Hit Rate', value: '94%' },
      { label: 'Enrichment Time', value: '2ms' },
      { label: 'Cache Size', value: '1,024' },
    ],
    recentActivity: [
      { time: '1s ago', message: 'Event enriched with pod metadata', type: 'success' },
      { time: '4s ago', message: 'Cache miss - queried K8s API', type: 'info' },
      { time: '12s ago', message: 'Pod cache updated', type: 'info' },
    ],
    dependencies: ['NATS', 'Kubernetes API'],
    config: [
      { key: 'Input', value: 'events.raw.*' },
      { key: 'Output', value: 'events.enriched' },
      { key: 'Cache TTL', value: '5m' },
    ],
  },
  {
    name: 'Detection',
    id: 'detect',
    status: 'running',
    description: 'Evaluates CEL rules against events, detects attacks, generates alerts',
    metrics: [
      { label: 'Events/sec', value: '~150' },
      { label: 'Rules Loaded', value: '5' },
      { label: 'Eval Time p95', value: '0.8ms' },
      { label: 'Alerts Today', value: '12' },
    ],
    recentActivity: [
      { time: '30s ago', message: 'ALERT: Shell Spawn in Prod', type: 'warning' },
      { time: '2m ago', message: 'ALERT: Privilege Escalation', type: 'error' },
      { time: '5m ago', message: 'Rule evaluation completed', type: 'info' },
    ],
    dependencies: ['NATS', 'Redis'],
    config: [
      { key: 'Input', value: 'events.enriched' },
      { key: 'Output', value: 'alerts' },
      { key: 'Correlation Window', value: '60s' },
    ],
  },
  {
    name: 'Incidents',
    id: 'incident',
    status: 'running',
    description: 'Groups alerts into incidents, maintains timeline, provides API for UI',
    metrics: [
      { label: 'Open Incidents', value: '3' },
      { label: 'Alerts Today', value: '12' },
      { label: 'API Latency', value: '8ms' },
      { label: 'DB Connections', value: '5' },
    ],
    recentActivity: [
      { time: '30s ago', message: 'Alert added to incident INC-001', type: 'info' },
      { time: '2m ago', message: 'New incident created: INC-003', type: 'warning' },
      { time: '10m ago', message: 'Incident INC-002 resolved', type: 'success' },
    ],
    dependencies: ['NATS', 'PostgreSQL'],
    config: [
      { key: 'API Port', value: '8081' },
      { key: 'Grouping Window', value: '1h' },
      { key: 'Auto-resolve', value: '24h' },
    ],
  },
  {
    name: 'Response',
    id: 'respond',
    status: 'running',
    description: 'Executes automated response playbooks with safety guardrails',
    metrics: [
      { label: 'Actions Today', value: '8' },
      { label: 'Success Rate', value: '100%' },
      { label: 'Blocked', value: '2' },
      { label: 'Avg Duration', value: '150ms' },
    ],
    recentActivity: [
      { time: '30s ago', message: 'Pod terminated: prod/vuln-nginx', type: 'success' },
      { time: '2m ago', message: 'Action blocked: kube-system protected', type: 'warning' },
      { time: '5m ago', message: 'Evidence bundle created', type: 'info' },
    ],
    dependencies: ['NATS', 'PostgreSQL', 'Kubernetes API'],
    config: [
      { key: 'Protected NS', value: 'kube-system, security-system' },
      { key: 'Playbooks', value: 'kill_pod, quarantine, isolate' },
      { key: 'Dry Run', value: 'false' },
    ],
  },
]

const statusColors = {
  running: 'bg-cyber-green',
  degraded: 'bg-severity-medium',
  error: 'bg-severity-critical',
  unknown: 'bg-gray-500',
}

const activityColors = {
  info: 'text-gray-400',
  success: 'text-cyber-green',
  warning: 'text-severity-medium',
  error: 'text-severity-critical',
}

export default function SystemComponents() {
  const [selectedComponent, setSelectedComponent] = useState<ComponentStatus | null>(null)

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="font-display text-lg font-semibold text-white">
          System Components
        </h3>
        <span className="text-xs text-gray-500">Click component for details</span>
      </div>

      <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-3">
        {componentData.map((component) => (
          <button
            key={component.id}
            onClick={() => setSelectedComponent(
              selectedComponent?.id === component.id ? null : component
            )}
            className={`flex flex-col items-center gap-2 p-4 rounded-lg transition-all text-left ${
              selectedComponent?.id === component.id
                ? 'bg-cyber-green/10 border border-cyber-green/30'
                : 'bg-midnight-700 hover:bg-midnight-600 border border-transparent'
            }`}
          >
            <div className="flex items-center gap-2 w-full">
              <span className={`w-2 h-2 rounded-full ${statusColors[component.status]} ${
                component.status === 'running' ? 'animate-pulse' : ''
              }`} />
              <span className="text-sm text-gray-300 truncate">{component.name}</span>
            </div>
            <div className="text-xs text-gray-500 capitalize">{component.status}</div>
          </button>
        ))}
      </div>

      <AnimatePresence>
        {selectedComponent && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            className="overflow-hidden"
          >
            <div className="bg-midnight-700 rounded-lg border border-midnight-600 p-6 mt-4">
              <div className="flex items-start justify-between mb-4">
                <div>
                  <div className="flex items-center gap-3">
                    <span className={`w-3 h-3 rounded-full ${statusColors[selectedComponent.status]}`} />
                    <h4 className="text-lg font-semibold text-white">{selectedComponent.name}</h4>
                    <span className={`px-2 py-0.5 rounded text-xs capitalize ${
                      selectedComponent.status === 'running' 
                        ? 'bg-cyber-green/20 text-cyber-green'
                        : selectedComponent.status === 'degraded'
                        ? 'bg-severity-medium/20 text-severity-medium'
                        : 'bg-severity-critical/20 text-severity-critical'
                    }`}>
                      {selectedComponent.status}
                    </span>
                  </div>
                  <p className="text-gray-400 text-sm mt-1">{selectedComponent.description}</p>
                </div>
                <button
                  onClick={() => setSelectedComponent(null)}
                  className="text-gray-500 hover:text-white p-1"
                >
                  x
                </button>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                {/* Metrics */}
                <div>
                  <h5 className="text-sm font-medium text-gray-400 mb-3 uppercase tracking-wide">Metrics</h5>
                  <div className="space-y-2">
                    {selectedComponent.metrics.map((metric, i) => (
                      <div key={i} className="flex justify-between items-center bg-midnight-800 rounded px-3 py-2">
                        <span className="text-gray-400 text-sm">{metric.label}</span>
                        <span className="text-white font-mono text-sm">{metric.value}</span>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Recent Activity */}
                <div>
                  <h5 className="text-sm font-medium text-gray-400 mb-3 uppercase tracking-wide">Recent Activity</h5>
                  <div className="space-y-2">
                    {selectedComponent.recentActivity.map((activity, i) => (
                      <div key={i} className="bg-midnight-800 rounded px-3 py-2">
                        <div className="flex justify-between items-start">
                          <span className={`text-sm ${activityColors[activity.type]}`}>
                            {activity.message}
                          </span>
                        </div>
                        <span className="text-xs text-gray-600">{activity.time}</span>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Configuration */}
                <div>
                  <h5 className="text-sm font-medium text-gray-400 mb-3 uppercase tracking-wide">Configuration</h5>
                  <div className="space-y-2">
                    {selectedComponent.config.map((cfg, i) => (
                      <div key={i} className="bg-midnight-800 rounded px-3 py-2">
                        <div className="text-gray-500 text-xs">{cfg.key}</div>
                        <div className="text-white text-sm font-mono truncate">{cfg.value}</div>
                      </div>
                    ))}
                  </div>
                  
                  <h5 className="text-sm font-medium text-gray-400 mt-4 mb-2 uppercase tracking-wide">Dependencies</h5>
                  <div className="flex flex-wrap gap-2">
                    {selectedComponent.dependencies.map((dep, i) => (
                      <span key={i} className="px-2 py-1 bg-midnight-800 rounded text-xs text-gray-400">
                        {dep}
                      </span>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}
