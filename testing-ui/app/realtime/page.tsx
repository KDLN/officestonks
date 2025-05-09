"use client"

import { useEffect, useState } from "react"
import Link from "next/link"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
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
import { stockWebSocket, StockUpdate } from "@/lib/websocket"

export default function RealtimePage() {
  const [stocks, setStocks] = useState<Stock[]>([])
  const [updates, setUpdates] = useState<StockUpdate[]>([])
  const [wsConnected, setWsConnected] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")

  // Load initial stocks
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

  // Set up WebSocket connection
  useEffect(() => {
    // Connect to WebSocket
    stockWebSocket.connect()
    
    // Add event listeners
    const onConnectUnsubscribe = stockWebSocket.onConnect(() => {
      setWsConnected(true)
    })
    
    const onDisconnectUnsubscribe = stockWebSocket.onDisconnect(() => {
      setWsConnected(false)
    })
    
    const onMessageUnsubscribe = stockWebSocket.onMessage((update) => {
      // Update the stocks array with the new price
      setStocks(prevStocks => 
        prevStocks.map(stock => 
          stock.id === update.id 
            ? { ...stock, current_price: update.current_price } 
            : stock
        )
      )
      
      // Add the update to the updates array (limited to last 20)
      setUpdates(prevUpdates => {
        const newUpdates = [update, ...prevUpdates]
        return newUpdates.slice(0, 20)
      })
    })
    
    // Clean up on unmount
    return () => {
      onConnectUnsubscribe()
      onDisconnectUnsubscribe()
      onMessageUnsubscribe()
      stockWebSocket.disconnect()
    }
  }, [])

  // Helper function to determine price change class
  const getPriceChangeClass = (update: StockUpdate) => {
    if (update.current_price > update.previous_price) {
      return "text-green-600"
    } else if (update.current_price < update.previous_price) {
      return "text-red-600"
    }
    return ""
  }

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-3xl font-bold mb-8">Real-time Stock Updates</h1>
        <p>Loading stocks...</p>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold">Real-time Stock Updates</h1>
        <div className="space-x-2">
          <Button asChild variant="outline">
            <Link href="/stocks">View All Stocks</Link>
          </Button>
          <Button asChild>
            <Link href="/portfolio">My Portfolio</Link>
          </Button>
        </div>
      </div>

      <div className="mb-6">
        <Card>
          <CardContent className="p-6">
            <div className="flex items-center">
              <div
                className={`h-3 w-3 rounded-full mr-2 ${
                  wsConnected ? "bg-green-500" : "bg-red-500"
                }`}
              ></div>
              <p>
                WebSocket Status:{" "}
                <span className="font-medium">
                  {wsConnected ? "Connected" : "Disconnected"}
                </span>
              </p>
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        <Card>
          <CardHeader>
            <CardTitle>Current Stock Prices</CardTitle>
            <CardDescription>
              Real-time stock prices from the Office Stonks market
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableCaption>
                Stock prices update in real-time via WebSocket
              </TableCaption>
              <TableHeader>
                <TableRow>
                  <TableHead>Symbol</TableHead>
                  <TableHead>Name</TableHead>
                  <TableHead className="text-right">Price</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {stocks.map((stock) => (
                  <TableRow key={stock.id}>
                    <TableCell className="font-bold">{stock.symbol}</TableCell>
                    <TableCell>{stock.name}</TableCell>
                    <TableCell className="text-right">
                      ${stock.current_price.toFixed(2)}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Recent Price Updates</CardTitle>
            <CardDescription>
              Latest price changes from the Office Stonks market
            </CardDescription>
          </CardHeader>
          <CardContent>
            {updates.length === 0 ? (
              <div className="text-center p-8">
                <p className="text-muted-foreground">
                  No price updates received yet. Updates will appear here in
                  real-time as they arrive.
                </p>
              </div>
            ) : (
              <Table>
                <TableCaption>
                  Most recent stock price updates (last 20)
                </TableCaption>
                <TableHeader>
                  <TableRow>
                    <TableHead>Time</TableHead>
                    <TableHead>Stock</TableHead>
                    <TableHead>Previous</TableHead>
                    <TableHead>Current</TableHead>
                    <TableHead>Change</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {updates.map((update, index) => {
                    const priceChangeClass = getPriceChangeClass(update)
                    const changeAmount = update.current_price - update.previous_price
                    const changePercent = (changeAmount / update.previous_price) * 100
                    
                    return (
                      <TableRow key={`${update.id}-${index}`}>
                        <TableCell>
                          {new Date(update.timestamp).toLocaleTimeString()}
                        </TableCell>
                        <TableCell className="font-bold">
                          {update.symbol}
                        </TableCell>
                        <TableCell>
                          ${update.previous_price.toFixed(2)}
                        </TableCell>
                        <TableCell className={priceChangeClass}>
                          ${update.current_price.toFixed(2)}
                        </TableCell>
                        <TableCell className={priceChangeClass}>
                          {changeAmount > 0 ? "+" : ""}
                          {changeAmount.toFixed(2)} (
                          {changeAmount > 0 ? "+" : ""}
                          {changePercent.toFixed(2)}%)
                        </TableCell>
                      </TableRow>
                    )
                  })}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}