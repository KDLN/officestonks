"use client"

import { useEffect, useState } from "react"
import { useRouter, useSearchParams } from "next/navigation"
import Link from "next/link"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { api, Stock, User, TradeRequest } from "@/lib/api"

export default function TradingPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  
  // Get stock_id and action from URL query params
  const stockIdParam = searchParams.get("stock_id")
  const actionParam = searchParams.get("action") || "buy"
  
  const [stocks, setStocks] = useState<Stock[]>([])
  const [selectedStockId, setSelectedStockId] = useState<number | null>(
    stockIdParam ? parseInt(stockIdParam, 10) : null
  )
  const [selectedStock, setSelectedStock] = useState<Stock | null>(null)
  const [transactionType, setTransactionType] = useState<"buy" | "sell">(
    actionParam === "sell" ? "sell" : "buy"
  )
  const [quantity, setQuantity] = useState<number>(1)
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState("")
  const [success, setSuccess] = useState("")

  // Load user data from localStorage
  useEffect(() => {
    const userStr = localStorage.getItem("user")
    if (userStr) {
      try {
        setUser(JSON.parse(userStr))
      } catch (err) {
        console.error("Failed to parse user data:", err)
      }
    }
  }, [])

  // Load all stocks
  useEffect(() => {
    const fetchStocks = async () => {
      try {
        const data = await api.getAllStocks()
        setStocks(data)
        
        // If a stock_id was provided in the URL, find that stock
        if (stockIdParam) {
          const stockId = parseInt(stockIdParam, 10)
          const stock = data.find(s => s.id === stockId)
          if (stock) {
            setSelectedStock(stock)
            setSelectedStockId(stockId)
          }
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load stocks")
        console.error(err)
      } finally {
        setLoading(false)
      }
    }

    fetchStocks()
  }, [stockIdParam])

  // Update selected stock when selectedStockId changes
  useEffect(() => {
    if (selectedStockId) {
      const stock = stocks.find(s => s.id === selectedStockId)
      if (stock) {
        setSelectedStock(stock)
      }
    } else {
      setSelectedStock(null)
    }
  }, [selectedStockId, stocks])

  // Calculate total cost
  const totalCost = selectedStock 
    ? quantity * selectedStock.current_price 
    : 0

  // Check if user has enough cash for buy transaction
  const hasEnoughCash = user && transactionType === "buy" 
    ? user.cash_balance >= totalCost 
    : true

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (!selectedStockId) {
      setError("Please select a stock")
      return
    }
    
    if (quantity <= 0) {
      setError("Quantity must be greater than 0")
      return
    }
    
    if (transactionType === "buy" && !hasEnoughCash) {
      setError("Insufficient funds for this transaction")
      return
    }
    
    setSubmitting(true)
    setError("")
    setSuccess("")
    
    try {
      const tradeData: TradeRequest = {
        stock_id: selectedStockId,
        quantity,
        transaction_type: transactionType
      }
      
      const result = await api.tradeStock(tradeData)
      
      // Update user's cash balance in localStorage
      if (user && result.user) {
        const updatedUser = { ...user, cash_balance: result.user.cash_balance }
        localStorage.setItem("user", JSON.stringify(updatedUser))
        setUser(updatedUser)
      }
      
      setSuccess(`Successfully ${transactionType === "buy" ? "purchased" : "sold"} ${quantity} shares of ${selectedStock?.symbol}`)
      
      // Reset form
      setQuantity(1)
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to complete transaction")
      console.error(err)
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-3xl font-bold mb-8">Trade Stocks</h1>
        <p>Loading stocks...</p>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold">Trade Stocks</h1>
        <div className="space-x-2">
          <Button asChild variant="outline">
            <Link href="/stocks">View All Stocks</Link>
          </Button>
          <Button asChild>
            <Link href="/portfolio">My Portfolio</Link>
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <div className="lg:col-span-2">
          <Card>
            <CardHeader>
              <CardTitle>Place an Order</CardTitle>
              <CardDescription>
                Buy or sell stocks in your portfolio
              </CardDescription>
            </CardHeader>
            <CardContent>
              <form onSubmit={handleSubmit} className="space-y-6">
                <div className="space-y-2">
                  <Label htmlFor="transaction-type">Transaction Type</Label>
                  <Select
                    value={transactionType}
                    onValueChange={(value) => setTransactionType(value as "buy" | "sell")}
                  >
                    <SelectTrigger id="transaction-type">
                      <SelectValue placeholder="Select transaction type" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="buy">Buy</SelectItem>
                      <SelectItem value="sell">Sell</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="stock-select">Select Stock</Label>
                  <Select
                    value={selectedStockId?.toString() || ""}
                    onValueChange={(value) => setSelectedStockId(parseInt(value, 10))}
                  >
                    <SelectTrigger id="stock-select">
                      <SelectValue placeholder="Select a stock" />
                    </SelectTrigger>
                    <SelectContent>
                      {stocks.map((stock) => (
                        <SelectItem key={stock.id} value={stock.id.toString()}>
                          {stock.symbol} - {stock.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                {selectedStock && (
                  <div className="p-4 bg-muted rounded-md">
                    <div className="grid grid-cols-2 gap-4">
                      <div>
                        <span className="text-sm text-muted-foreground">
                          Stock
                        </span>
                        <p className="font-medium">
                          {selectedStock.symbol} - {selectedStock.name}
                        </p>
                      </div>
                      <div>
                        <span className="text-sm text-muted-foreground">
                          Current Price
                        </span>
                        <p className="font-medium">
                          ${selectedStock.current_price.toFixed(2)}
                        </p>
                      </div>
                    </div>
                  </div>
                )}

                <div className="space-y-2">
                  <Label htmlFor="quantity">Quantity</Label>
                  <Input
                    id="quantity"
                    type="number"
                    min="1"
                    value={quantity}
                    onChange={(e) => setQuantity(parseInt(e.target.value, 10) || 0)}
                    required
                  />
                </div>

                {selectedStock && (
                  <div className="p-4 bg-muted rounded-md">
                    <div className="flex justify-between">
                      <span className="text-sm font-medium">Total Cost</span>
                      <span className="font-bold">
                        ${totalCost.toFixed(2)}
                      </span>
                    </div>
                  </div>
                )}

                {error && (
                  <div className="p-4 bg-red-50 text-red-800 rounded-md">
                    {error}
                  </div>
                )}

                {success && (
                  <div className="p-4 bg-green-50 text-green-800 rounded-md">
                    {success}
                  </div>
                )}

                <Button
                  type="submit"
                  className="w-full"
                  disabled={
                    submitting ||
                    !selectedStockId ||
                    quantity <= 0 ||
                    (transactionType === "buy" && !hasEnoughCash)
                  }
                >
                  {submitting
                    ? "Processing..."
                    : `${
                        transactionType === "buy" ? "Buy" : "Sell"
                      } ${quantity} Shares for $${totalCost.toFixed(2)}`}
                </Button>
              </form>
            </CardContent>
          </Card>
        </div>

        {user && (
          <div>
            <Card>
              <CardHeader>
                <CardTitle>Account Information</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div>
                    <span className="text-sm text-muted-foreground">
                      Username
                    </span>
                    <p className="font-medium">{user.username}</p>
                  </div>
                  <div>
                    <span className="text-sm text-muted-foreground">
                      Cash Balance
                    </span>
                    <p className="text-xl font-bold">
                      ${user.cash_balance.toFixed(2)}
                    </p>
                  </div>
                  {transactionType === "buy" && selectedStock && (
                    <div>
                      <span className="text-sm text-muted-foreground">
                        Balance After Transaction
                      </span>
                      <p
                        className={`text-xl font-bold ${
                          hasEnoughCash ? "text-green-600" : "text-red-600"
                        }`}
                      >
                        ${(user.cash_balance - totalCost).toFixed(2)}
                      </p>
                    </div>
                  )}
                </div>
              </CardContent>
              <CardFooter>
                <Button asChild variant="outline" className="w-full">
                  <Link href="/portfolio">View Portfolio</Link>
                </Button>
              </CardFooter>
            </Card>
          </div>
        )}
      </div>
    </div>
  )
}