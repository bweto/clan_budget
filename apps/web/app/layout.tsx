import React from 'react'
import './globals.css'

export const metadata = {
  title: 'Clan Budget',
  description: 'Manage your family budget with ease.',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  )
}
