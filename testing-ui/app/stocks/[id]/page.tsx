"use client"

import { useEffect, useState } from "react"
import { useRouter } from "next/navigation"
import Link from "next/link"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card"
import { api, Stock } from "@/lib/api"

export default function StockDetailPage({ params }: { params: { id: string } }) {
  const router = useRouter()
  const [stock, setStock] = useState<Stock | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")

  useEffect(() => {
    const fetchStock = async () => {
      try {
        const data = await api.getStockById(parseInt(params.id, 10))
        setStock(data)
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load stock details")
        console.error(err)
      } finally {
        setLoading(false)
      }
    }

    fetchStock()
  }, [params.id])

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Button variant="outline" asChild className="mb-4">
          <Link href="/stocks">← Back to Stocks</Link>
        </Button>
        <p>Loading stock details...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Button variant="outline" asChild className="mb-4">
          <Link href="/stocks">← Back to Stocks</Link>
        </Button>
        <Card>
          <CardContent className="p-6">
            <p className="text-red-500">{error}</p>
            <Button className="mt-4" onClick={() => window.location.reload()}>
              Try Again
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  if (!stock) {
    return (
      <div className="container mx-auto px-4 py-8">
        <Button variant="outline" asChild className="mb-4">
          <Link href="/stocks">← Back to Stocks</Link>
        </Button>
        <Card>
          <CardContent className="p-6">
            <p>Stock not found</p>
          </CardContent>
        </Card>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <Button variant="outline" asChild className="mb-4">
        <Link href="/stocks">← Back to Stocks</Link>
      </Button>
      
      <Card className="w-full max-w-3xl mx-auto">
        <CardHeader>
          <CardTitle className="text-3xl">
            {stock.symbol} - {stock.name}
          </CardTitle>
          <CardDescription>{stock.sector}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex justify-between items-center p-4 bg-muted rounded-md">
            <span className="text-muted-foreground">Current Price</span>
            <span className="text-2xl font-bold">${stock.current_price.toFixed(2)}</span>
          </div>
          
          <div className="grid grid-cols-2 gap-4">
            <div className="p-4 bg-muted rounded-md">
              <span className="text-muted-foreground">Stock ID</span>
              <p>{stock.id}</p>
            </div>
            <div className="p-4 bg-muted rounded-md">
              <span className="text-muted-foreground">Last Updated</span>
              <p>{new Date(stock.last_updated).toLocaleString()}</p>
            </div>
          </div>
        </CardContent>
        <CardFooter>
          <Button asChild className="w-full">
            <Link href={`/trading?stock_id=${stock.id}`}>Trade This Stock</Link>
          </Button>
        </CardFooter>
      </Card>
    </div>
  )
}