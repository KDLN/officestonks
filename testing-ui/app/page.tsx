"use client"

import { useState } from 'react'
import Link from 'next/link'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card'

export default function Home() {
  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold">Office Stonks Testing UI</h1>
        <div className="space-x-2">
          <Button asChild variant="outline">
            <Link href="/login">Login</Link>
          </Button>
          <Button asChild>
            <Link href="/register">Register</Link>
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        <FeatureCard 
          title="Authentication"
          description="Register accounts and login to test the authentication system."
          link="/auth"
        />
        
        <FeatureCard 
          title="Stock Market"
          description="View all available stocks and detailed information about each one."
          link="/stocks"
        />
        
        <FeatureCard 
          title="Portfolio"
          description="View your personal portfolio of stocks and transaction history."
          link="/portfolio"
        />
        
        <FeatureCard 
          title="Trading"
          description="Buy and sell stocks to build your portfolio."
          link="/trading"
        />
        
        <FeatureCard 
          title="Real-time Updates"
          description="See real-time updates to stock prices via WebSocket."
          link="/realtime"
        />
        
        <FeatureCard 
          title="API Documentation"
          description="View the available API endpoints for Office Stonks."
          link="/api-docs"
        />
      </div>
    </div>
  )
}

function FeatureCard({ title, description, link }: { title: string, description: string, link: string }) {
  return (
    <Card className="h-full">
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      <CardFooter>
        <Button asChild>
          <Link href={link}>Try it</Link>
        </Button>
      </CardFooter>
    </Card>
  )
}