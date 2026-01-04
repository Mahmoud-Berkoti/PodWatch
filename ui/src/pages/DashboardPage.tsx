import { useQuery } from '@tanstack/react-query'
import { motion } from 'framer-motion'
import axios from 'axios'
import SystemComponents from '../components/SystemComponents'

interface Stats {
  totalAlerts: number
  criticalAlerts: number
  openIncidents: number
  actionsExecuted: number
}

async function fetchAlerts() {
  const { data } = await axios.get('/api/alerts')
  return data || []
}

async function fetchIncidents() {
  const { data } = await axios.get('/api/incidents')
  return data || []
}

export default function DashboardPage() {
  const { data: alerts = [] } = useQuery({ queryKey: ['alerts'], queryFn: fetchAlerts })
  const { data: incidents = [] } = useQuery({ queryKey: ['incidents'], queryFn: fetchIncidents })

  const stats: Stats = {
    totalAlerts: alerts.length,
    criticalAlerts: alerts.filter((a: any) => a.severity === 'critical').length,
    openIncidents: incidents.filter((i: any) => i.status === 'open').length,
    actionsExecuted: incidents.length * 2, // Placeholder
  }

  const statCards = [
    { label: 'Total Alerts', value: stats.totalAlerts, icon: 'A', color: 'cyber-blue' },
    { label: 'Critical', value: stats.criticalAlerts, icon: 'C', color: 'severity-critical' },
    { label: 'Open Incidents', value: stats.openIncidents, icon: 'I', color: 'severity-high' },
    { label: 'Actions Taken', value: stats.actionsExecuted, icon: 'R', color: 'cyber-green' },
  ]

  return (
    <div className="space-y-8">
      <div>
        <h2 className="font-display text-2xl font-bold text-white mb-2">
          Security Dashboard
        </h2>
        <p className="text-gray-500">Real-time threat detection and response</p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {statCards.map((stat, index) => (
          <motion.div
            key={stat.label}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: index * 0.1 }}
            className="bg-midnight-800 rounded-xl p-6 border border-midnight-600 hover:border-midnight-500 transition-all"
          >
            <div className="flex items-center justify-between mb-4">
              <span className={`w-10 h-10 rounded-lg bg-${stat.color}/20 text-${stat.color} flex items-center justify-center font-bold`}>
                {stat.icon}
              </span>
              <span className={`text-${stat.color} text-3xl font-bold font-display`}>
                {stat.value}
              </span>
            </div>
            <div className="text-gray-400 text-sm">{stat.label}</div>
          </motion.div>
        ))}
      </div>

      {/* Activity Feed */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <motion.div
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: 0.4 }}
          className="bg-midnight-800 rounded-xl p-6 border border-midnight-600"
        >
          <div className="flex items-center justify-between mb-4">
            <h3 className="font-display text-lg font-semibold text-white">
              Recent Alerts
            </h3>
            <a href="/alerts" className="text-xs text-cyber-green hover:underline">
              View all
            </a>
          </div>
          <div className="space-y-3 max-h-[400px] overflow-y-auto pr-2">
            {alerts.slice(0, 10).map((alert: any) => (
              <div
                key={alert.id}
                className="p-4 bg-midnight-700 rounded-lg hover:bg-midnight-600 transition-colors cursor-pointer"
              >
                <div className="flex items-start gap-3">
                  <span
                    className={`w-2 h-2 rounded-full mt-2 flex-shrink-0 ${
                      alert.severity === 'critical'
                        ? 'bg-severity-critical animate-pulse'
                        : alert.severity === 'high'
                        ? 'bg-severity-high'
                        : 'bg-severity-medium'
                    }`}
                  />
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between gap-2 mb-1">
                      <span className="text-white text-sm font-medium truncate">
                        {alert.rule_name}
                      </span>
                      <span
                        className={`px-2 py-0.5 rounded text-xs font-bold flex-shrink-0 ${
                          alert.severity === 'critical'
                            ? 'bg-severity-critical/20 text-severity-critical'
                            : alert.severity === 'high'
                            ? 'bg-severity-high/20 text-severity-high'
                            : 'bg-severity-medium/20 text-severity-medium'
                        }`}
                      >
                        {alert.severity}
                      </span>
                    </div>
                    <p className="text-gray-400 text-xs mb-2 line-clamp-2">
                      {alert.description || 'Security alert triggered by runtime detection'}
                    </p>
                    <div className="flex flex-wrap gap-2 text-xs">
                      {alert.namespace && (
                        <span className="px-2 py-0.5 bg-midnight-800 rounded text-gray-400">
                          ns: {alert.namespace}
                        </span>
                      )}
                      {alert.pod_name && (
                        <span className="px-2 py-0.5 bg-midnight-800 rounded text-gray-400">
                          pod: {alert.pod_name}
                        </span>
                      )}
                      {alert.response && (
                        <span className="px-2 py-0.5 bg-cyber-green/10 rounded text-cyber-green">
                          {alert.response}
                        </span>
                      )}
                    </div>
                    <div className="flex items-center justify-between mt-2 pt-2 border-t border-midnight-600">
                      <span className="text-gray-500 text-xs">
                        {alert.created_at
                          ? new Date(alert.created_at).toLocaleString()
                          : 'Just now'}
                      </span>
                      <span className="text-gray-600 text-xs font-mono">
                        {alert.id?.slice(0, 8) || 'N/A'}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            ))}
            {alerts.length === 0 && (
              <div className="text-center py-8 text-gray-500">
                <div className="mb-2">No recent alerts</div>
                <p className="text-xs">Alerts will appear here when threats are detected</p>
              </div>
            )}
          </div>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: 0.5 }}
          className="bg-midnight-800 rounded-xl p-6 border border-midnight-600"
        >
          <h3 className="font-display text-lg font-semibold text-white mb-4">
            Active Incidents
          </h3>
          <div className="space-y-3">
            {incidents
              .filter((i: any) => i.status === 'open' || i.status === 'investigating')
              .slice(0, 5)
              .map((incident: any) => (
                <a
                  key={incident.id}
                  href={`/incidents/${incident.id}`}
                  className="flex items-center gap-3 p-3 bg-midnight-700 rounded-lg hover:bg-midnight-600 transition-colors"
                >
                  <span className="w-2 h-2 rounded-full bg-severity-critical" />
                  <div className="flex-1 min-w-0">
                    <div className="text-white text-sm font-medium truncate">
                      {incident.title}
                    </div>
                    <div className="text-gray-500 text-xs capitalize">
                      {incident.status}
                    </div>
                  </div>
                  <span
                    className={`px-2 py-1 rounded text-xs font-bold ${
                      incident.severity === 'critical'
                        ? 'bg-severity-critical/20 text-severity-critical'
                        : 'bg-severity-high/20 text-severity-high'
                    }`}
                  >
                    {incident.severity}
                  </span>
                </a>
              ))}
            {incidents.filter((i: any) => i.status === 'open').length === 0 && (
              <div className="text-center py-8 text-gray-500">
                <p>No active incidents</p>
              </div>
            )}
          </div>
        </motion.div>
      </div>

      {/* System Components - Interactive */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.6 }}
        className="bg-midnight-800 rounded-xl p-6 border border-midnight-600"
      >
        <SystemComponents />
      </motion.div>

      {/* Alert Logs - Terminal Style */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.7 }}
        className="bg-midnight-900 rounded-xl border border-midnight-600 overflow-hidden"
      >
        <div className="flex items-center justify-between px-4 py-3 bg-midnight-800 border-b border-midnight-600">
          <div className="flex items-center gap-3">
            <div className="flex gap-1.5">
              <span className="w-3 h-3 rounded-full bg-severity-critical" />
              <span className="w-3 h-3 rounded-full bg-severity-medium" />
              <span className="w-3 h-3 rounded-full bg-severity-low" />
            </div>
            <h3 className="font-display text-sm font-semibold text-white">
              Alert Logs
            </h3>
          </div>
          <div className="flex items-center gap-2">
            <span className="text-xs text-gray-500">{alerts.length} entries</span>
            <a href="/alerts" className="text-xs text-cyber-green hover:underline">
              Full View
            </a>
          </div>
        </div>
        <div className="p-4 font-mono text-sm max-h-[350px] overflow-y-auto">
          {alerts.length > 0 ? (
            <div className="space-y-1">
              {alerts.slice(0, 20).map((alert: any, index: number) => {
                const timestamp = alert.created_at
                  ? new Date(alert.created_at).toISOString().replace('T', ' ').slice(0, 19)
                  : new Date().toISOString().replace('T', ' ').slice(0, 19)
                const severityColor =
                  alert.severity === 'critical'
                    ? 'text-severity-critical'
                    : alert.severity === 'high'
                    ? 'text-severity-high'
                    : 'text-severity-medium'
                return (
                  <div
                    key={alert.id || index}
                    className="flex flex-wrap gap-x-2 py-1 hover:bg-midnight-800 px-2 -mx-2 rounded cursor-pointer group"
                  >
                    <span className="text-gray-600">{timestamp}</span>
                    <span className={`font-bold uppercase w-20 ${severityColor}`}>
                      [{alert.severity || 'INFO'}]
                    </span>
                    <span className="text-cyber-blue">
                      {alert.namespace || 'default'}/{alert.pod_name || 'unknown'}
                    </span>
                    <span className="text-gray-400">-</span>
                    <span className="text-white flex-1">
                      {alert.rule_name || 'Alert'}
                    </span>
                    {alert.response && (
                      <span className="text-cyber-green">
                        [{alert.response}]
                      </span>
                    )}
                    <span className="text-gray-700 group-hover:text-gray-500 text-xs self-center">
                      {alert.id?.slice(0, 8) || ''}
                    </span>
                  </div>
                )
              })}
            </div>
          ) : (
            <div className="text-gray-600 text-center py-8">
              <p>-- No alert logs available --</p>
              <p className="text-xs mt-2">Logs will stream here as alerts are generated</p>
            </div>
          )}
        </div>
        <div className="px-4 py-2 bg-midnight-800 border-t border-midnight-600 flex items-center justify-between">
          <div className="flex items-center gap-4 text-xs">
            <span className="flex items-center gap-1">
              <span className="w-2 h-2 rounded-full bg-severity-critical" />
              <span className="text-gray-500">Critical: {alerts.filter((a: any) => a.severity === 'critical').length}</span>
            </span>
            <span className="flex items-center gap-1">
              <span className="w-2 h-2 rounded-full bg-severity-high" />
              <span className="text-gray-500">High: {alerts.filter((a: any) => a.severity === 'high').length}</span>
            </span>
            <span className="flex items-center gap-1">
              <span className="w-2 h-2 rounded-full bg-severity-medium" />
              <span className="text-gray-500">Medium: {alerts.filter((a: any) => a.severity === 'medium').length}</span>
            </span>
          </div>
          <span className="text-xs text-gray-600">
            Last updated: {new Date().toLocaleTimeString()}
          </span>
        </div>
      </motion.div>
    </div>
  )
}
