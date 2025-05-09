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
import { api, Stock } from "@/lib/api"

export default function StocksPage() {
  const [stocks, setStocks] = useState<Stock[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")

  useEffect(() => {
    const fetchStocks = async () => {
      try {
        const data = await api.getAllStocks()
        setStocks(data)
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load stocks")
        console.error(err)
      } finally {
        setLoading(false)
      }
    }

    fetchStocks()
  }, [])

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-3xl font-bold mb-8">Stock Market</h1>
        <p>Loading stocks...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-3xl font-bold mb-8">Stock Market</h1>
        <Card>
          <CardContent className="p-6">
            <p className="text-red-500">{error}</p>
            <p className="mt-4">
              This might be because you&apos;re not connected to the API server or it&apos;s not running.
            </p>
            <Button className="mt-4" onClick={() => window.location.reload()}>
              Try Again
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold">Stock Market</h1>
        <div className="space-x-2">
          <Button asChild variant="outline">
            <Link href="/portfolio">My Portfolio</Link>
          </Button>
          <Button asChild>
            <Link href="/trading">Trade Stocks</Link>
          </Button>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Available Stocks</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableCaption>
              A list of all available stocks for trading on Office Stonks.
            </TableCaption>
            <TableHeader>
              <TableRow>
                <TableHead>Symbol</TableHead>
                <TableHead>Name</TableHead>
                <TableHead>Sector</TableHead>
                <TableHead className="text-right">Price</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {stocks.map((stock) => (
                <TableRow key={stock.id}>
                  <TableCell className="font-bold">{stock.symbol}</TableCell>
                  <TableCell>{stock.name}</TableCell>
                  <TableCell>{stock.sector}</TableCell>
                  <TableCell className="text-right">
                    ${stock.current_price.toFixed(2)}
                  </TableCell>
                  <TableCell className="text-right">
                    <Button
                      asChild
                      variant="outline"
                      size="sm"
                      className="mr-2"
                    >
                      <Link href={`/stocks/${stock.id}`}>Details</Link>
                    </Button>
                    <Button asChild size="sm">
                      <Link href={`/trading?stock_id=${stock.id}`}>Trade</Link>
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}