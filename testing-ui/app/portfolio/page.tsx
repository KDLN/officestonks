"use client"

import { useEffect, useState } from "react"
import Link from "next/link"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { api, PortfolioItem, User } from "@/lib/api"

export default function PortfolioPage() {
  const [portfolio, setPortfolio] = useState<PortfolioItem[]>([])
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")

  useEffect(() => {
    // Get user data from localStorage
    const userStr = localStorage.getItem("user")
    if (userStr) {
      try {
        setUser(JSON.parse(userStr))
      } catch (err) {
        console.error("Failed to parse user data:", err)
      }
    }

    const fetchPortfolio = async () => {
      try {
        const data = await api.getUserPortfolio()
        setPortfolio(data)
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load portfolio")
        console.error(err)
      } finally {
        setLoading(false)
      }
    }

    fetchPortfolio()
  }, [])

  // Calculate the total portfolio value
  const totalValue = portfolio.reduce(
    (sum, item) => sum + item.quantity * item.stock.current_price,
    0
  )

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-3xl font-bold mb-8">My Portfolio</h1>
        <p>Loading portfolio...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-3xl font-bold mb-8">My Portfolio</h1>
        <Card>
          <CardContent className="p-6">
            <p className="text-red-500">{error}</p>
            <p className="mt-4">
              This might be because you&apos;re not logged in or the API server isn&apos;t running.
            </p>
            <div className="flex gap-2 mt-4">
              <Button asChild variant="outline">
                <Link href="/login">Login</Link>
              </Button>
              <Button onClick={() => window.location.reload()}>
                Try Again
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold">My Portfolio</h1>
        <div className="space-x-2">
          <Button asChild variant="outline">
            <Link href="/stocks">View All Stocks</Link>
          </Button>
          <Button asChild>
            <Link href="/transactions">Transaction History</Link>
          </Button>
        </div>
      </div>

      {user && (
        <Card className="mb-8">
          <CardContent className="p-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="p-4 bg-muted rounded-md">
                <span className="text-muted-foreground">Username</span>
                <p className="text-xl font-bold">{user.username}</p>
              </div>
              <div className="p-4 bg-muted rounded-md">
                <span className="text-muted-foreground">Cash Balance</span>
                <p className="text-xl font-bold">${user.cash_balance.toFixed(2)}</p>
              </div>
              <div className="p-4 bg-muted rounded-md">
                <span className="text-muted-foreground">Portfolio Value</span>
                <p className="text-xl font-bold">${totalValue.toFixed(2)}</p>
              </div>
              <div className="p-4 bg-muted rounded-md">
                <span className="text-muted-foreground">Total Assets</span>
                <p className="text-xl font-bold">
                  ${(user.cash_balance + totalValue).toFixed(2)}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Your Stock Holdings</CardTitle>
        </CardHeader>
        <CardContent>
          {portfolio.length === 0 ? (
            <div className="text-center p-8">
              <p className="text-lg text-muted-foreground mb-4">
                You don&apos;t own any stocks yet.
              </p>
              <Button asChild>
                <Link href="/stocks">Browse Stocks to Buy</Link>
              </Button>
            </div>
          ) : (
            <Table>
              <TableCaption>
                Your current stock holdings in Office Stonks.
              </TableCaption>
              <TableHeader>
                <TableRow>
                  <TableHead>Symbol</TableHead>
                  <TableHead>Name</TableHead>
                  <TableHead className="text-right">Quantity</TableHead>
                  <TableHead className="text-right">Price Per Share</TableHead>
                  <TableHead className="text-right">Total Value</TableHead>
                  <TableHead className="text-right">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {portfolio.map((item) => (
                  <TableRow key={item.id}>
                    <TableCell className="font-bold">
                      {item.stock.symbol}
                    </TableCell>
                    <TableCell>{item.stock.name}</TableCell>
                    <TableCell className="text-right">{item.quantity}</TableCell>
                    <TableCell className="text-right">
                      ${item.stock.current_price.toFixed(2)}
                    </TableCell>
                    <TableCell className="text-right">
                      ${(item.quantity * item.stock.current_price).toFixed(2)}
                    </TableCell>
                    <TableCell className="text-right">
                      <Button
                        asChild
                        variant="outline"
                        size="sm"
                        className="mr-2"
                      >
                        <Link href={`/stocks/${item.stock_id}`}>Details</Link>
                      </Button>
                      <Button asChild size="sm">
                        <Link
                          href={`/trading?stock_id=${item.stock_id}&action=sell`}
                        >
                          Sell
                        </Link>
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  )
}