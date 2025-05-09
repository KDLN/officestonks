"use client"

import Link from "next/link"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"

export default function ApiDocsPage() {
  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <h1 className="text-3xl font-bold">API Documentation</h1>
        <Button asChild variant="outline">
          <Link href="/">Back to Home</Link>
        </Button>
      </div>

      <div className="space-y-10">
        <Card>
          <CardHeader>
            <CardTitle>Overview</CardTitle>
            <CardDescription>
              Office Stonks provides a RESTful API for interacting with the stock trading platform
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <p>
              The API is available at <code className="bg-muted px-1 py-0.5 rounded">http://your-app-url</code>. All endpoints return JSON responses and accept JSON for POST/PUT requests.
            </p>
            <p>
              Protected endpoints require authentication via JWT tokens, which should be included in the <code className="bg-muted px-1 py-0.5 rounded">Authorization</code> header as <code className="bg-muted px-1 py-0.5 rounded">Bearer &lt;token&gt;</code>.
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Authentication Endpoints</CardTitle>
          </CardHeader>
          <CardContent className="space-y-8">
            <div className="space-y-4">
              <div className="flex items-center">
                <span className="bg-blue-100 text-blue-800 px-2 py-1 rounded-md mr-3 font-mono text-sm">POST</span>
                <code className="font-bold">/api/auth/register</code>
              </div>
              <p className="text-muted-foreground">Register a new user account</p>
              <div className="space-y-2">
                <p className="font-medium">Request Body:</p>
                <pre className="bg-muted p-4 rounded-md overflow-x-auto">
                  {`{
  "username": "string",
  "password": "string"
}`}
                </pre>
              </div>
              <div className="space-y-2">
                <p className="font-medium">Response:</p>
                <pre className="bg-muted p-4 rounded-md overflow-x-auto">
                  {`{
  "token": "string",
  "user": {
    "id": "number",
    "username": "string",
    "cash_balance": "number"
  }
}`}
                </pre>
              </div>
            </div>

            <div className="space-y-4">
              <div className="flex items-center">
                <span className="bg-blue-100 text-blue-800 px-2 py-1 rounded-md mr-3 font-mono text-sm">POST</span>
                <code className="font-bold">/api/auth/login</code>
              </div>
              <p className="text-muted-foreground">Login with existing credentials</p>
              <div className="space-y-2">
                <p className="font-medium">Request Body:</p>
                <pre className="bg-muted p-4 rounded-md overflow-x-auto">
                  {`{
  "username": "string",
  "password": "string"
}`}
                </pre>
              </div>
              <div className="space-y-2">
                <p className="font-medium">Response:</p>
                <pre className="bg-muted p-4 rounded-md overflow-x-auto">
                  {`{
  "token": "string",
  "user": {
    "id": "number",
    "username": "string",
    "cash_balance": "number"
  }
}`}
                </pre>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Stock Endpoints</CardTitle>
          </CardHeader>
          <CardContent className="space-y-8">
            <div className="space-y-4">
              <div className="flex items-center">
                <span className="bg-green-100 text-green-800 px-2 py-1 rounded-md mr-3 font-mono text-sm">GET</span>
                <code className="font-bold">/api/stocks</code>
              </div>
              <p className="text-muted-foreground">Get a list of all available stocks</p>
              <div className="space-y-2">
                <p className="font-medium">Response:</p>
                <pre className="bg-muted p-4 rounded-md overflow-x-auto">
                  {`[
  {
    "id": "number",
    "symbol": "string",
    "name": "string",
    "sector": "string",
    "current_price": "number",
    "last_updated": "string"
  }
]`}
                </pre>
              </div>
            </div>

            <div className="space-y-4">
              <div className="flex items-center">
                <span className="bg-green-100 text-green-800 px-2 py-1 rounded-md mr-3 font-mono text-sm">GET</span>
                <code className="font-bold">/api/stocks/{"{id}"}</code>
              </div>
              <p className="text-muted-foreground">Get details for a specific stock</p>
              <div className="space-y-2">
                <p className="font-medium">Response:</p>
                <pre className="bg-muted p-4 rounded-md overflow-x-auto">
                  {`{
  "id": "number",
  "symbol": "string",
  "name": "string",
  "sector": "string",
  "current_price": "number",
  "last_updated": "string"
}`}
                </pre>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Portfolio Endpoints</CardTitle>
            <CardDescription>Requires authentication</CardDescription>
          </CardHeader>
          <CardContent className="space-y-8">
            <div className="space-y-4">
              <div className="flex items-center">
                <span className="bg-green-100 text-green-800 px-2 py-1 rounded-md mr-3 font-mono text-sm">GET</span>
                <code className="font-bold">/api/portfolio</code>
              </div>
              <p className="text-muted-foreground">Get the current user's portfolio</p>
              <div className="space-y-2">
                <p className="font-medium">Response:</p>
                <pre className="bg-muted p-4 rounded-md overflow-x-auto">
                  {`[
  {
    "id": "number",
    "user_id": "number",
    "stock_id": "number",
    "quantity": "number",
    "stock": {
      "id": "number",
      "symbol": "string",
      "name": "string",
      "sector": "string",
      "current_price": "number",
      "last_updated": "string"
    }
  }
]`}
                </pre>
              </div>
            </div>

            <div className="space-y-4">
              <div className="flex items-center">
                <span className="bg-green-100 text-green-800 px-2 py-1 rounded-md mr-3 font-mono text-sm">GET</span>
                <code className="font-bold">/api/transactions</code>
              </div>
              <p className="text-muted-foreground">Get the current user's transaction history</p>
              <div className="space-y-2">
                <p className="font-medium">Response:</p>
                <pre className="bg-muted p-4 rounded-md overflow-x-auto">
                  {`[
  {
    "id": "number",
    "user_id": "number",
    "stock_id": "number",
    "quantity": "number",
    "price": "number",
    "transaction_type": "buy" | "sell",
    "created_at": "string",
    "stock": {
      "id": "number",
      "symbol": "string",
      "name": "string"
    }
  }
]`}
                </pre>
              </div>
            </div>

            <div className="space-y-4">
              <div className="flex items-center">
                <span className="bg-blue-100 text-blue-800 px-2 py-1 rounded-md mr-3 font-mono text-sm">POST</span>
                <code className="font-bold">/api/trading</code>
              </div>
              <p className="text-muted-foreground">Buy or sell stocks</p>
              <div className="space-y-2">
                <p className="font-medium">Request Body:</p>
                <pre className="bg-muted p-4 rounded-md overflow-x-auto">
                  {`{
  "stock_id": "number",
  "quantity": "number",
  "transaction_type": "buy" | "sell"
}`}
                </pre>
              </div>
              <div className="space-y-2">
                <p className="font-medium">Response:</p>
                <pre className="bg-muted p-4 rounded-md overflow-x-auto">
                  {`{
  "transaction": {
    "id": "number",
    "user_id": "number",
    "stock_id": "number",
    "quantity": "number",
    "price": "number",
    "transaction_type": "buy" | "sell",
    "created_at": "string"
  },
  "user": {
    "id": "number",
    "username": "string",
    "cash_balance": "number"
  }
}`}
                </pre>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>WebSocket API</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p>
              Office Stonks provides a WebSocket connection for real-time stock price updates at <code className="bg-muted px-1 py-0.5 rounded">ws://your-app-url/ws</code>.
            </p>
            
            <div className="space-y-2">
              <p className="font-medium">Message Format:</p>
              <pre className="bg-muted p-4 rounded-md overflow-x-auto">
                {`{
  "id": "number",
  "symbol": "string",
  "current_price": "number",
  "previous_price": "number",
  "timestamp": "string"
}`}
              </pre>
            </div>
            
            <p>
              No authentication is required for WebSocket connections. Connect to receive real-time price updates for all stocks.
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}