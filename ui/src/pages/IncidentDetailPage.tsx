import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { motion } from 'framer-motion'
import axios from 'axios'
import IncidentTimeline from '../components/IncidentTimeline'

async function fetchIncident(id: string) {
  const { data } = await axios.get(`/api/incidents/${id}`)
  return data
}

async function fetchTimeline(id: string) {
  const { data } = await axios.get(`/api/incidents/${id}/timeline`)
  return data || []
}

async function updateIncidentStatus(id: string, status: string) {
  const { data } = await axios.patch(`/api/incidents/${id}`, { status })
  return data
}

const statusOptions = ['open', 'investigating', 'contained', 'resolved']

export default function IncidentDetailPage() {
  const { id } = useParams<{ id: string }>()
  const queryClient = useQueryClient()

  const { data: incident, isLoading } = useQuery({
    queryKey: ['incident', id],
    queryFn: () => fetchIncident(id!),
    enabled: !!id,
  })

  const { data: timeline = [] } = useQuery({
    queryKey: ['incident-timeline', id],
    queryFn: () => fetchTimeline(id!),
    enabled: !!id,
  })

  const mutation = useMutation({
    mutationFn: (status: string) => updateIncidentStatus(id!, status),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['incident', id] })
    },
  })

  if (isLoading || !incident) {
    return (
      <div className="text-center py-12 text-gray-500">
        <div className="animate-spin w-8 h-8 border-2 border-cyber-green border-t-transparent rounded-full mx-auto mb-4" />
        Loading incident...
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        className="bg-midnight-800 rounded-xl p-6 border border-midnight-600"
      >
        <div className="flex items-start justify-between">
          <div>
            <div className="flex items-center gap-3 mb-3">
              <span
                className={`px-3 py-1 rounded text-sm font-bold uppercase ${
                  incident.severity === 'critical'
                    ? 'bg-severity-critical text-white'
                    : incident.severity === 'high'
                    ? 'bg-severity-high text-black'
                    : 'bg-severity-medium text-black'
                }`}
              >
                {incident.severity}
              </span>
              <span className="text-gray-500">•</span>
              <span className="text-gray-400 font-mono text-sm">{incident.id}</span>
            </div>
            <h1 className="font-display text-2xl font-bold text-white mb-2">
              {incident.title}
            </h1>
            <p className="text-gray-500">
              Created: {new Date(incident.created_at).toLocaleString()}
            </p>
          </div>

          <div className="flex flex-col items-end gap-3">
            <div className="text-sm text-gray-400">Status</div>
            <div className="flex gap-2">
              {statusOptions.map((status) => (
                <button
                  key={status}
                  onClick={() => mutation.mutate(status)}
                  disabled={mutation.isPending}
                  className={`px-3 py-1.5 rounded text-sm font-medium capitalize transition-all ${
                    incident.status === status
                      ? status === 'resolved'
                        ? 'bg-cyber-green text-black'
                        : status === 'contained'
                        ? 'bg-severity-medium text-black'
                        : status === 'investigating'
                        ? 'bg-severity-high text-black'
                        : 'bg-severity-critical text-white'
                      : 'bg-midnight-700 text-gray-400 hover:text-white'
                  }`}
                >
                  {status}
                </button>
              ))}
            </div>
          </div>
        </div>
      </motion.div>

      {/* Timeline */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.1 }}
        className="bg-midnight-800 rounded-xl p-6 border border-midnight-600"
      >
        <h2 className="font-display text-lg font-semibold text-white mb-6">
          Incident Timeline
        </h2>
        <IncidentTimeline events={timeline} />
      </motion.div>
    </div>
  )
}
