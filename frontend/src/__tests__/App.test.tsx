import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import App from '../App'
import { UiProvider } from '../contexts/UiContext'

describe('App', () => {
  it('renders without crashing', () => {
    render(
      <UiProvider>
        <App />
      </UiProvider>
    )
    expect(screen.getByText('Distributed Job Scheduler')).toBeInTheDocument()
  })
})