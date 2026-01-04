import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { motion } from 'framer-motion'
import axios from 'axios'
import AlertTable from '../components/AlertTable'

async function fetchAlerts(severity?: string) {
  const params = severity ? { severity } : {}
  const { data } = await axios.get('/api/alerts', { params })
  return data || []
}

export default function AlertsPage() {
  const [severityFilter, setSeverityFilter] = useState<string>('')
  const { data: alerts = [], isLoading } = useQuery({
    queryKey: ['alerts', severityFilter],
    queryFn: () => fetchAlerts(severityFilter),
  })

  const severities = ['', 'critical', 'high', 'medium', 'low']

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="font-display text-2xl font-bold text-white">Alerts</h2>
          <p className="text-gray-500">Runtime threat detections</p>
        </div>

        {/* Filters */}
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <span className="text-gray-400 text-sm">Severity:</span>
            <div className="flex gap-1">
              {severities.map((sev) => (
                <button
                  key={sev || 'all'}
                  onClick={() => setSeverityFilter(sev)}
                  className={`px-3 py-1.5 rounded text-sm font-medium transition-all ${
                    severityFilter === sev
                      ? 'bg-cyber-green/20 text-cyber-green border border-cyber-green/30'
                      : 'bg-midnight-700 text-gray-400 hover:text-white'
                  }`}
                >
                  {sev || 'All'}
                </button>
              ))}
            </div>
          </div>
        </div>
      </div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="bg-midnight-800 rounded-xl p-6 border border-midnight-600"
      >
        {isLoading ? (
          <div className="text-center py-12 text-gray-500">
            <div className="animate-spin w-8 h-8 border-2 border-cyber-green border-t-transparent rounded-full mx-auto mb-4" />
            Loading alerts...
          </div>
        ) : (
          <AlertTable alerts={alerts} />
        )}
      </motion.div>
    </div>
  )
}
