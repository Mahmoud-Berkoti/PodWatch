import { Outlet, NavLink } from 'react-router-dom'
import { motion } from 'framer-motion'

const navItems = [
  { path: '/dashboard', label: 'Dashboard', icon: '' },
  { path: '/alerts', label: 'Alerts', icon: '' },
  { path: '/incidents', label: 'Incidents', icon: '' },
  { path: '/rules', label: 'Rules', icon: '' },
  { path: '/replay', label: 'Replay', icon: '' },
]

export default function Layout() {
  return (
    <div className="min-h-screen bg-midnight-900 grid-bg">
      {/* Header */}
      <header className="border-b border-midnight-600 bg-midnight-800/80 backdrop-blur-sm sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <motion.div
              initial={{ rotate: -10 }}
              animate={{ rotate: 0 }}
              transition={{ type: 'spring', stiffness: 200 }}
            >
              <img src="/shield.svg" alt="KubeGuard" className="w-10 h-10" />
            </motion.div>
            <div>
              <h1 className="font-display text-xl font-bold text-cyber-green text-glow">
                PodWatch
              </h1>
              <p className="text-xs text-gray-500">Runtime Threat Detection</p>
            </div>
          </div>
          
          <nav className="flex gap-1">
            {navItems.map((item) => (
              <NavLink
                key={item.path}
                to={item.path}
                className={({ isActive }) =>
                  `px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200 ${
                    isActive
                      ? 'bg-cyber-green/10 text-cyber-green border border-cyber-green/30'
                      : 'text-gray-400 hover:text-white hover:bg-midnight-700'
                  }`
                }
              >
                {item.label}
              </NavLink>
            ))}
          </nav>

          <div className="flex items-center gap-4">
            <div className="flex items-center gap-2 text-sm">
              <span className="w-2 h-2 rounded-full bg-cyber-green animate-pulse" />
              <span className="text-gray-400">Connected</span>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-6 py-8">
        <Outlet />
      </main>
    </div>
  )
}
