import { useState } from 'react'
import { motion } from 'framer-motion'

const defaultRules = `# KubeGuard Detection Rules
# YAML format with CEL expressions

rules:
  - id: "rule-shell-spawn"
    name: "Shell Spawn in Production"
    description: "Interactive shell spawned in production namespace"
    severity: "high"
    condition: |
      event.process.exe in ['/bin/bash', '/bin/sh'] && 
      event.container.namespace == 'prod'
    response: "kill_pod"
    enabled: true

  - id: "rule-token-read"
    name: "Service Account Token Read"
    description: "Process reading Kubernetes service account token"
    severity: "high"
    condition: |
      event.event_type == 'file_open' && 
      event.process.cmdline.contains('/var/run/secrets')
    response: "quarantine_namespace"
    enabled: true

  - id: "rule-reverse-shell"
    name: "Reverse Shell Indicators"
    description: "Shell process connecting to external IP"
    severity: "critical"
    condition: |
      event.event_type == 'network_connect' && 
      event.process.exe.endsWith('bash') &&
      !event.network.dst_ip.startsWith('10.')
    response: "kill_pod"
    enabled: true

  - id: "rule-priv-esc"
    name: "Privilege Escalation"
    description: "Container gained sensitive capabilities"
    severity: "critical"
    condition: |
      event.process.capabilities_added.exists(c, c == 'SYS_ADMIN')
    response: "isolate_node"
    enabled: true

  - id: "rule-pkg-manager"
    name: "Package Manager in Production"
    description: "Package manager executed in prod"
    severity: "medium"
    condition: |
      event.process.exe in ['/usr/bin/apt', '/sbin/apk'] &&
      event.container.namespace == 'prod'
    response: ""
    enabled: true
`

export default function RulesPage() {
  const [rules, setRules] = useState(defaultRules)
  const [saved, setSaved] = useState(false)

  const handleSave = () => {
    // In a real app, this would POST to the API
    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="font-display text-2xl font-bold text-white">Rules</h2>
          <p className="text-gray-500">Detection rule configuration</p>
        </div>
        <motion.button
          whileHover={{ scale: 1.02 }}
          whileTap={{ scale: 0.98 }}
          onClick={handleSave}
          className={`px-6 py-2 rounded-lg font-medium transition-all ${
            saved
              ? 'bg-cyber-green text-black'
              : 'bg-cyber-green/20 text-cyber-green border border-cyber-green/30 hover:bg-cyber-green/30'
          }`}
        >
          {saved ? 'Saved' : 'Save Rules'}
        </motion.button>
      </div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="bg-midnight-800 rounded-xl border border-midnight-600 overflow-hidden"
      >
        {/* Editor Header */}
        <div className="flex items-center justify-between px-4 py-3 bg-midnight-700 border-b border-midnight-600">
          <div className="flex items-center gap-3">
            <div className="flex gap-1.5">
              <span className="w-3 h-3 rounded-full bg-severity-critical" />
              <span className="w-3 h-3 rounded-full bg-severity-medium" />
              <span className="w-3 h-3 rounded-full bg-cyber-green" />
            </div>
            <span className="text-gray-400 text-sm font-mono">rules.yaml</span>
          </div>
          <div className="text-gray-500 text-xs">
            {rules.split('\n').length} lines
          </div>
        </div>

        {/* Editor */}
        <div className="relative">
          <div className="absolute left-0 top-0 bottom-0 w-12 bg-midnight-700/50 flex flex-col items-end pr-3 pt-4 text-gray-600 text-sm font-mono select-none">
            {rules.split('\n').map((_, i) => (
              <div key={i} className="leading-6">
                {i + 1}
              </div>
            ))}
          </div>
          <textarea
            value={rules}
            onChange={(e) => setRules(e.target.value)}
            className="w-full min-h-[600px] bg-transparent text-gray-300 font-mono text-sm p-4 pl-16 resize-none focus:outline-none leading-6"
            spellCheck={false}
          />
        </div>
      </motion.div>

      {/* Help Section */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.1 }}
        className="bg-midnight-800 rounded-xl p-6 border border-midnight-600"
      >
        <h3 className="font-display text-lg font-semibold text-white mb-4">
          CEL Expression Reference
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
          <div className="bg-midnight-700 rounded-lg p-4">
            <h4 className="text-cyber-green font-medium mb-2">Event Fields</h4>
            <ul className="space-y-1 text-gray-400 font-mono text-xs">
              <li>event.process.exe</li>
              <li>event.process.cmdline</li>
              <li>event.container.namespace</li>
              <li>event.container.pod</li>
              <li>event.network.dst_ip</li>
              <li>event.network.dst_port</li>
            </ul>
          </div>
          <div className="bg-midnight-700 rounded-lg p-4">
            <h4 className="text-cyber-blue font-medium mb-2">Response Actions</h4>
            <ul className="space-y-1 text-gray-400 font-mono text-xs">
              <li>kill_pod</li>
              <li>quarantine_namespace</li>
              <li>isolate_node</li>
              <li>evidence_bundle</li>
              <li>"" (alert only)</li>
            </ul>
          </div>
        </div>
      </motion.div>
    </div>
  )
}
