import React from 'react'
import { createRoot } from 'react-dom/client'
import App from './App'
import './styles.css'

const rootElement = document.getElementById('presentation-root')
if (rootElement) {
  createRoot(rootElement).render(
    <React.StrictMode>
      <App />
    </React.StrictMode>,
  )
}
