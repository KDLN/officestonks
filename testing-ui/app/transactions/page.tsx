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
import { api, Transaction } from "@/lib/api"

export default function TransactionsPage() {
  const [transactions, setTransactions] = useState<Transaction[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")

  useEffect(() => {
    const fetchTransactions = async () => {
      try {
        const data = await api.getTransactionHistory()
        setTransactions(data)
      } catch (err) {
        setError(
          err instanceof Error
            ? err.message
            : "Failed to load transaction history"
        )
        console.error(err)
      } finally {
        setLoading(false)
      }
    }

    fetchTransactions()
  }, [])

  if (loading) {
    return (
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-3xl font-bold mb-8">Transaction History</h1>
        <p>Loading transactions...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-3xl font-bold mb-8">Transaction History</h1>
        <Card>
          <CardContent className="p-6">
            <p className="text-red-500">{error}</p>
            <p className="mt-4">
              This might be because you&apos;re not logged in or the API server isn&apos;t
              running.
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
        <h1 className="text-3xl font-bold">Transaction History</h1>
        <div className="space-x-2">
          <Button asChild variant="outline">
            <Link href="/stocks">View All Stocks</Link>
          </Button>
          <Button asChild>
            <Link href="/portfolio">My Portfolio</Link>
          </Button>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Your Trading History</CardTitle>
        </CardHeader>
        <CardContent>
          {transactions.length === 0 ? (
            <div className="text-center p-8">
              <p className="text-lg text-muted-foreground mb-4">
                You haven&apos;t made any transactions yet.
              </p>
              <Button asChild>
                <Link href="/stocks">Browse Stocks to Trade</Link>
              </Button>
            </div>
          ) : (
            <Table>
              <TableCaption>
                A history of all your transactions on Office Stonks.
              </TableCaption>
              <TableHeader>
                <TableRow>
                  <TableHead>Date</TableHead>
                  <TableHead>Stock</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead className="text-right">Quantity</TableHead>
                  <TableHead className="text-right">Price Per Share</TableHead>
                  <TableHead className="text-right">Total Value</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {transactions.map((transaction) => (
                  <TableRow key={transaction.id}>
                    <TableCell>
                      {new Date(transaction.created_at).toLocaleString()}
                    </TableCell>
                    <TableCell>
                      <div className="font-bold">{transaction.stock.symbol}</div>
                      <div className="text-sm text-muted-foreground">
                        {transaction.stock.name}
                      </div>
                    </TableCell>
                    <TableCell>
                      <span
                        className={`px-2 py-1 rounded-full text-xs font-medium ${
                          transaction.transaction_type === "buy"
                            ? "bg-green-100 text-green-800"
                            : "bg-red-100 text-red-800"
                        }`}
                      >
                        {transaction.transaction_type.toUpperCase()}
                      </span>
                    </TableCell>
                    <TableCell className="text-right">
                      {transaction.quantity}
                    </TableCell>
                    <TableCell className="text-right">
                      ${transaction.price.toFixed(2)}
                    </TableCell>
                    <TableCell className="text-right">
                      ${(transaction.quantity * transaction.price).toFixed(2)}
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