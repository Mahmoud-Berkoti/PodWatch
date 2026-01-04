import { motion } from 'framer-motion'
import { formatDistanceToNow } from 'date-fns'

interface Alert {
  id: string
  timestamp: string
  rule_name: string
  severity: string
  description: string
  incident_id: string
  response: string
}

interface AlertTableProps {
  alerts: Alert[]
  onAlertClick?: (alert: Alert) => void
}

const severityColors: Record<string, string> = {
  critical: 'bg-severity-critical text-white',
  high: 'bg-severity-high text-black',
  medium: 'bg-severity-medium text-black',
  low: 'bg-severity-low text-black',
  info: 'bg-severity-info text-black',
}

const severityGlow: Record<string, string> = {
  critical: 'shadow-[0_0_15px_rgba(255,45,85,0.4)]',
  high: 'shadow-[0_0_15px_rgba(255,149,0,0.4)]',
  medium: 'shadow-[0_0_15px_rgba(255,204,0,0.3)]',
  low: '',
  info: '',
}

export default function AlertTable({ alerts, onAlertClick }: AlertTableProps) {
  if (!alerts || alerts.length === 0) {
    return (
      <div className="text-center py-12 text-gray-500">
        <p className="text-lg font-medium mb-2">No Alerts</p>
        <p>No alerts detected. System secure.</p>
      </div>
    )
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full">
        <thead>
          <tr className="text-left text-sm text-gray-500 border-b border-midnight-600">
            <th className="pb-3 font-medium">Severity</th>
            <th className="pb-3 font-medium">Rule</th>
            <th className="pb-3 font-medium">Description</th>
            <th className="pb-3 font-medium">Time</th>
            <th className="pb-3 font-medium">Response</th>
            <th className="pb-3 font-medium">Incident</th>
          </tr>
        </thead>
        <tbody>
          {alerts.map((alert, index) => (
            <motion.tr
              key={alert.id}
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: index * 0.05 }}
              onClick={() => onAlertClick?.(alert)}
              className="border-b border-midnight-700 hover:bg-midnight-700/50 cursor-pointer transition-colors"
            >
              <td className="py-4">
                <span
                  className={`px-2 py-1 rounded text-xs font-bold uppercase ${
                    severityColors[alert.severity] || 'bg-gray-600'
                  } ${severityGlow[alert.severity] || ''}`}
                >
                  {alert.severity}
                </span>
              </td>
              <td className="py-4 font-medium text-white">{alert.rule_name}</td>
              <td className="py-4 text-gray-400 max-w-md truncate">
                {alert.description}
              </td>
              <td className="py-4 text-gray-500 text-sm">
                {formatDistanceToNow(new Date(alert.timestamp), { addSuffix: true })}
              </td>
              <td className="py-4">
                {alert.response ? (
                  <span className="px-2 py-1 bg-cyber-purple/20 text-cyber-purple rounded text-xs">
                    {alert.response}
                  </span>
                ) : (
                  <span className="text-gray-600">—</span>
                )}
              </td>
              <td className="py-4">
                {alert.incident_id ? (
                  <a
                    href={`/incidents/${alert.incident_id}`}
                    className="text-cyber-blue hover:underline text-sm"
                    onClick={(e) => e.stopPropagation()}
                  >
                    {alert.incident_id.slice(0, 8)}...
                  </a>
                ) : (
                  <span className="text-gray-600">—</span>
                )}
              </td>
            </motion.tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
