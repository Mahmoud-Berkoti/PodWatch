import { motion } from 'framer-motion'
import { formatDistanceToNow, format } from 'date-fns'

interface TimelineEvent {
  type: 'alert' | 'action'
  id: string
  timestamp: string
  rule_name?: string
  severity?: string
  description?: string
  action_type?: string
  target?: string
  status?: string
  message?: string
  event?: {
    process?: {
      exe: string
      cmdline: string
      pid: number
    }
    container?: {
      namespace: string
      pod: string
      image: string
    }
    network?: {
      dst_ip: string
      dst_port: number
    }
  }
}

interface IncidentTimelineProps {
  events: TimelineEvent[]
}

export default function IncidentTimeline({ events }: IncidentTimelineProps) {
  if (!events || events.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500">
        No timeline events yet.
      </div>
    )
  }

  return (
    <div className="relative">
      {/* Vertical line */}
      <div className="absolute left-6 top-0 bottom-0 w-px bg-gradient-to-b from-cyber-green via-cyber-blue to-transparent" />

      <div className="space-y-6">
        {events.map((event, index) => (
          <motion.div
            key={event.id}
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ delay: index * 0.1 }}
            className="relative pl-16"
          >
            {/* Timeline dot */}
            <div
              className={`absolute left-4 w-5 h-5 rounded-full border-2 ${
                event.type === 'alert'
                  ? 'bg-severity-critical border-severity-critical'
                  : event.status === 'success'
                  ? 'bg-cyber-green border-cyber-green'
                  : event.status === 'blocked'
                  ? 'bg-severity-medium border-severity-medium'
                  : 'bg-gray-500 border-gray-500'
              }`}
            />

            <div className="bg-midnight-700 rounded-lg p-4 border border-midnight-600 hover:border-midnight-500 transition-colors">
              {/* Header */}
              <div className="flex items-start justify-between mb-3">
                <div>
                  <span
                    className={`text-xs font-bold uppercase px-2 py-0.5 rounded ${
                      event.type === 'alert'
                        ? 'bg-severity-critical/20 text-severity-critical'
                        : 'bg-cyber-blue/20 text-cyber-blue'
                    }`}
                  >
                    {event.type}
                  </span>
                  <h4 className="text-white font-medium mt-2">
                    {event.type === 'alert' ? event.rule_name : event.action_type}
                  </h4>
                </div>
                <div className="text-right text-sm text-gray-500">
                  <div>{format(new Date(event.timestamp), 'HH:mm:ss')}</div>
                  <div>{formatDistanceToNow(new Date(event.timestamp), { addSuffix: true })}</div>
                </div>
              </div>

              {/* Content */}
              {event.type === 'alert' && event.event && (
                <div className="space-y-2 text-sm">
                  {event.event.process && (
                    <div className="bg-midnight-800 rounded p-2">
                      <div className="text-gray-400 text-xs mb-1">Process</div>
                      <div className="font-mono text-cyber-green">
                        {event.event.process.cmdline || event.event.process.exe}
                      </div>
                      <div className="text-gray-500 text-xs mt-1">
                        PID: {event.event.process.pid}
                      </div>
                    </div>
                  )}
                  {event.event.container && (
                    <div className="bg-midnight-800 rounded p-2">
                      <div className="text-gray-400 text-xs mb-1">Container</div>
                      <div className="text-white">
                        {event.event.container.namespace}/{event.event.container.pod}
                      </div>
                      <div className="text-gray-500 text-xs mt-1">
                        {event.event.container.image}
                      </div>
                    </div>
                  )}
                  {event.event.network && event.event.network.dst_ip && (
                    <div className="bg-midnight-800 rounded p-2">
                      <div className="text-gray-400 text-xs mb-1">Network</div>
                      <div className="font-mono text-cyber-orange">
                        {event.event.network.dst_ip}:{event.event.network.dst_port}
                      </div>
                    </div>
                  )}
                </div>
              )}

              {event.type === 'action' && (
                <div className="space-y-2 text-sm">
                  <div className="flex items-center gap-2">
                    <span className="text-gray-400">Target:</span>
                    <span className="font-mono text-white">{event.target}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-gray-400">Status:</span>
                    <span
                      className={`px-2 py-0.5 rounded text-xs font-medium ${
                        event.status === 'success'
                          ? 'bg-cyber-green/20 text-cyber-green'
                          : event.status === 'blocked'
                          ? 'bg-severity-medium/20 text-severity-medium'
                          : 'bg-severity-critical/20 text-severity-critical'
                      }`}
                    >
                      {event.status}
                    </span>
                  </div>
                  {event.message && (
                    <div className="text-gray-400">{event.message}</div>
                  )}
                </div>
              )}
            </div>
          </motion.div>
        ))}
      </div>
    </div>
  )
}
