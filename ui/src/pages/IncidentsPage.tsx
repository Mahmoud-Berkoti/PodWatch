import { useQuery } from '@tanstack/react-query'
import { motion } from 'framer-motion'
import { formatDistanceToNow } from 'date-fns'
import { Link } from 'react-router-dom'
import axios from 'axios'

interface Incident {
  id: string
  status: string
  severity: string
  title: string
  created_at: string
  updated_at: string
}

async function fetchIncidents() {
  const { data } = await axios.get('/api/incidents')
  return data || []
}

const statusColors: Record<string, string> = {
  open: 'bg-severity-critical/20 text-severity-critical border-severity-critical/30',
  investigating: 'bg-severity-high/20 text-severity-high border-severity-high/30',
  contained: 'bg-severity-medium/20 text-severity-medium border-severity-medium/30',
  resolved: 'bg-cyber-green/20 text-cyber-green border-cyber-green/30',
}

export default function IncidentsPage() {
  const { data: incidents = [], isLoading } = useQuery({
    queryKey: ['incidents'],
    queryFn: fetchIncidents,
  })

  return (
    <div className="space-y-6">
      <div>
        <h2 className="font-display text-2xl font-bold text-white">Incidents</h2>
        <p className="text-gray-500">Security incidents and response tracking</p>
      </div>

      {isLoading ? (
        <div className="text-center py-12 text-gray-500">
          <div className="animate-spin w-8 h-8 border-2 border-cyber-green border-t-transparent rounded-full mx-auto mb-4" />
          Loading incidents...
        </div>
      ) : incidents.length === 0 ? (
        <div className="text-center py-16 text-gray-500">
          <p className="text-xl font-medium">No incidents recorded</p>
          <p className="mt-2">Your cluster is secure</p>
        </div>
      ) : (
        <div className="grid gap-4">
          {incidents.map((incident: Incident, index: number) => (
            <motion.div
              key={incident.id}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.05 }}
            >
              <Link
                to={`/incidents/${incident.id}`}
                className="block bg-midnight-800 rounded-xl p-6 border border-midnight-600 hover:border-cyber-green/30 transition-all group"
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1">
                    <div className="flex items-center gap-3 mb-2">
                      <span
                        className={`px-2 py-1 rounded text-xs font-bold uppercase border ${
                          statusColors[incident.status] || 'bg-gray-600'
                        }`}
                      >
                        {incident.status}
                      </span>
                      <span
                        className={`px-2 py-1 rounded text-xs font-bold ${
                          incident.severity === 'critical'
                            ? 'bg-severity-critical text-white'
                            : incident.severity === 'high'
                            ? 'bg-severity-high text-black'
                            : 'bg-severity-medium text-black'
                        }`}
                      >
                        {incident.severity}
                      </span>
                    </div>
                    <h3 className="text-lg font-medium text-white group-hover:text-cyber-green transition-colors">
                      {incident.title}
                    </h3>
                    <p className="text-gray-500 text-sm mt-1">
                      ID: {incident.id.slice(0, 8)}...
                    </p>
                  </div>
                  <div className="text-right text-sm text-gray-500">
                    <div>Created {formatDistanceToNow(new Date(incident.created_at), { addSuffix: true })}</div>
                    <div>Updated {formatDistanceToNow(new Date(incident.updated_at), { addSuffix: true })}</div>
                  </div>
                </div>
              </Link>
            </motion.div>
          ))}
        </div>
      )}
    </div>
  )
}
