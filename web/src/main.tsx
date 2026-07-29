import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import App from './App'
import { applyAppearance, getPreferredLook, getPreferredTheme } from './theme'
import './styles.css'

applyAppearance(getPreferredLook(), getPreferredTheme())

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
