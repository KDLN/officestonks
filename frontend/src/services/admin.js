// Admin service for frontend
import { authenticatedFetch } from "./authBridge";

// Check the current hostname to determine if we're running locally
const isLocalhost =
  window.location.hostname === "localhost" ||
  window.location.hostname === "127.0.0.1";

// Get configuration from environment variables with fallbacks
const BACKEND_URL =
  process.env.REACT_APP_BACKEND_URL ||
  "https://officestonks.com";

// Connect directly to backend (no CORS proxy needed)
const BASE_URL = isLocalhost
  ? "/api" // Use relative URL when running locally
  : `${BACKEND_URL}/api`; // Direct connection to backend in production

const API_URL = BASE_URL.endsWith("/api") ? BASE_URL : `${BASE_URL}/api`;
const ADMIN_URL = `${API_URL}/admin`;

console.log("Admin service using API URL:", API_URL);

// Check if current user has admin privileges
export const checkAdminStatus = async () => {
  try {
    const response = await authenticatedFetch(`${API_URL}/admin/status`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to check admin status");
    }

    const data = await response.json();
    return data.isAdmin || false;
  } catch (error) {
    console.error("Error checking admin status:", error);
    throw error;
  }
};

// Get all users (admin only)
export const getAllUsers = async () => {
  try {
    const response = await authenticatedFetch(`${ADMIN_URL}/users`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to fetch users");
    }

    return await response.json();
  } catch (error) {
    console.error("Error fetching users:", error);
    throw error;
  }
};

// Update user (admin only)
export const updateUser = async (userId, updates) => {
  try {
    const response = await authenticatedFetch(`${ADMIN_URL}/users/${userId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(updates),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to update user");
    }

    return await response.json();
  } catch (error) {
    console.error("Error updating user:", error);
    throw error;
  }
};

// Delete user (admin only)
export const deleteUser = async (userId) => {
  try {
    const response = await authenticatedFetch(`${ADMIN_URL}/users/${userId}`, {
      method: "DELETE",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to delete user");
    }

    return await response.json();
  } catch (error) {
    console.error("Error deleting user:", error);
    throw error;
  }
};

// Stock Management APIs
export const getAllStocksDetailed = async () => {
  try {
    const response = await authenticatedFetch(`${ADMIN_URL}/stocks`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to fetch stocks");
    }

    return await response.json();
  } catch (error) {
    console.error("Error fetching stocks:", error);
    throw error;
  }
};

export const createStock = async (stockData) => {
  try {
    const response = await authenticatedFetch(`${ADMIN_URL}/stocks`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(stockData),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to create stock");
    }

    return await response.json();
  } catch (error) {
    console.error("Error creating stock:", error);
    throw error;
  }
};

export const updateStock = async (stockId, stockData) => {
  try {
    console.log(`📡 API Call: PUT ${ADMIN_URL}/stocks/${stockId}`, stockData);
    
    const response = await authenticatedFetch(`${ADMIN_URL}/stocks/${stockId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(stockData),
    });

    console.log(`📡 Response status: ${response.status} ${response.statusText}`);

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      console.error(`📡 Error response:`, errorData);
      throw new Error(errorData.error || `HTTP ${response.status}: Failed to update stock`);
    }

    const responseData = await response.json();
    console.log(`📡 Success response:`, responseData);
    return responseData;
  } catch (error) {
    console.error("❌ updateStock API error:", error);
    throw error;
  }
};

export const deleteStock = async (stockId) => {
  try {
    const response = await authenticatedFetch(`${ADMIN_URL}/stocks/${stockId}`, {
      method: "DELETE",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to delete stock");
    }

    return await response.json();
  } catch (error) {
    console.error("Error deleting stock:", error);
    throw error;
  }
};

export const launchIPO = async (ipoData) => {
  try {
    const response = await authenticatedFetch(`${ADMIN_URL}/stocks/ipo`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(ipoData),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to launch IPO");
    }

    return await response.json();
  } catch (error) {
    console.error("Error launching IPO:", error);
    throw error;
  }
};

export const triggerSectorEvent = async (eventData) => {
  try {
    const response = await authenticatedFetch(`${ADMIN_URL}/stocks/sector-event`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(eventData),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to trigger sector event");
    }

    return await response.json();
  } catch (error) {
    console.error("Error triggering sector event:", error);
    throw error;
  }
};

// Reset stock prices (admin only)
export const resetStockPrices = async () => {
  try {
    const response = await authenticatedFetch(`${ADMIN_URL}/stocks/reset`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to reset stock prices");
    }

    return await response.json();
  } catch (error) {
    console.error("Error resetting stock prices:", error);
    throw error;
  }
};

// Clear all chats (admin only)
export const clearAllChats = async () => {
  try {
    const response = await authenticatedFetch(`${ADMIN_URL}/chat/clear`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to clear chats");
    }

    return await response.json();
  } catch (error) {
    console.error("Error clearing chats:", error);
    throw error;
  }
};


// Game Configuration Management
export const getGameConfig = async () => {
  try {
    const response = await authenticatedFetch(`${ADMIN_URL}/game-config`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to fetch game configuration");
    }

    return await response.json();
  } catch (error) {
    console.error("Error fetching game config:", error);
    throw error;
  }
};

export const updateGameConfig = async (config) => {
  try {
    const response = await authenticatedFetch(`${ADMIN_URL}/game-config`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(config),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to update game configuration");
    }

    return await response.json();
  } catch (error) {
    console.error("Error updating game config:", error);
    throw error;
  }
};

export const resetGameConfig = async () => {
  try {
    const response = await authenticatedFetch(`${ADMIN_URL}/game-config/reset`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to reset game configuration");
    }

    return await response.json();
  } catch (error) {
    console.error("Error resetting game config:", error);
    throw error;
  }
};

export const loadBalancedConfig = async () => {
  try {
    const response = await authenticatedFetch(`${ADMIN_URL}/game-config/balanced`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to load balanced configuration");
    }

    return await response.json();
  } catch (error) {
    console.error("Error loading balanced config:", error);
    throw error;
  }
};

// Crisis Testing Functions

// Force a crisis event for a specific stock
export const forceCrisisEvent = async (stockId) => {
  try {
    const response = await authenticatedFetch(`${ADMIN_URL}/crisis/force`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ stock_id: stockId }),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to force crisis event");
    }

    return await response.json();
  } catch (error) {
    console.error("Error forcing crisis event:", error);
    throw error;
  }
};

// Force a bankruptcy event for a specific stock
export const forceBankruptcy = async (stockId) => {
  try {
    const response = await authenticatedFetch(`${ADMIN_URL}/crisis/bankruptcy`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ stock_id: stockId }),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to force bankruptcy");
    }

    return await response.json();
  } catch (error) {
    console.error("Error forcing bankruptcy:", error);
    throw error;
  }
};

// Force a recovery event for a specific stock
export const forceRecovery = async (stockId) => {
  try {
    const response = await authenticatedFetch(`${ADMIN_URL}/crisis/recovery`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ stock_id: stockId }),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to force recovery");
    }

    return await response.json();
  } catch (error) {
    console.error("Error forcing recovery:", error);
    throw error;
  }
};

// Get simulator status for all stocks
export const getSimulatorStatus = async () => {
  try {
    const response = await authenticatedFetch(`${ADMIN_URL}/simulator/status`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to get simulator status");
    }

    return await response.json();
  } catch (error) {
    console.error("Error getting simulator status:", error);
    throw error;
  }
};

// Test orchestration functions
export const getTestStatus = async () => {
  try {
    const response = await authenticatedFetch(`${ADMIN_URL}/tests/status`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to get test status");
    }

    return await response.json();
  } catch (error) {
    console.error("Error getting test status:", error);
    throw error;
  }
};

export const runCrisisTests = async () => {
  try {
    const response = await authenticatedFetch(`${ADMIN_URL}/tests/crisis`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to run crisis tests");
    }

    return await response.json();
  } catch (error) {
    console.error("Error running crisis tests:", error);
    throw error;
  }
};

export const runPortfolioTests = async () => {
  try {
    const response = await authenticatedFetch(`${ADMIN_URL}/tests/portfolio`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to run portfolio tests");
    }

    return await response.json();
  } catch (error) {
    console.error("Error running portfolio tests:", error);
    throw error;
  }
};

// Run SSE tests (admin only)
export const runSSETests = async () => {
  try {
    const response = await authenticatedFetch(`${ADMIN_URL}/tests/sse`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to run SSE tests");
    }

    return await response.json();
  } catch (error) {
    console.error("Error running SSE tests:", error);
    throw error;
  }
};

// Run stock management tests (admin only) - for debugging 500 errors
export const runStockManagementTests = async () => {
  try {
    const response = await authenticatedFetch(`${ADMIN_URL}/tests/stock-management`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to run stock management tests");
    }

    return await response.json();
  } catch (error) {
    console.error("Error running stock management tests:", error);
    throw error;
  }
};

// Create missing sectors to fix foreign key constraints (admin only)
export const createMissingSectors = async () => {
  try {
    const response = await authenticatedFetch(`${ADMIN_URL}/tests/create-sectors`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to create missing sectors");
    }

    return await response.json();
  } catch (error) {
    console.error("Error creating missing sectors:", error);
    throw error;
  }
};

// Test admin stock update endpoint (admin only) - for debugging edit issues
export const testAdminStockUpdate = async () => {
  try {
    const response = await authenticatedFetch(`${ADMIN_URL}/tests/admin-stock-update`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || "Failed to test admin stock update");
    }

    return await response.json();
  } catch (error) {
    console.error("Error testing admin stock update:", error);
    throw error;
  }
};

// Test stock price locking mechanism (admin only) - verify price stays locked for 30 seconds
export const testStockPriceLocking = async () => {
  try {
    console.log("🔒 Starting stock price locking test...");
    
    // First get a stock to test with (use first available stock)
    const stocksResponse = await authenticatedFetch(`${ADMIN_URL}/stocks`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!stocksResponse.ok) {
      throw new Error("Failed to fetch stocks for test");
    }

    const stocks = await stocksResponse.json();
    if (!stocks || stocks.length === 0) {
      throw new Error("No stocks available for testing");
    }

    const testStock = stocks[0]; // Use first stock
    const testStockId = testStock.id;
    const originalPrice = testStock.current_price;
    const testPrice = Math.round((originalPrice * 1.5) * 100) / 100; // 50% increase, rounded to 2 decimals

    console.log(`🔒 Testing with stock ${testStock.symbol} (ID: ${testStockId})`);
    console.log(`🔒 Original price: $${originalPrice}, Test price: $${testPrice}`);

    const testResults = {
      stock_id: testStockId,
      stock_symbol: testStock.symbol,
      original_price: originalPrice,
      test_price: testPrice,
      steps: [],
      success: false,
      error: null
    };

    try {
      // Step 1: Update stock price via admin endpoint
      testResults.steps.push({ step: 1, action: "Updating stock price", timestamp: new Date().toISOString() });
      
      await updateStock(testStockId, {
        current_price: testPrice
      });

      console.log(`🔒 Step 1 complete: Stock price updated to $${testPrice}`);
      testResults.steps.push({ 
        step: 1, 
        action: "Stock price updated successfully", 
        timestamp: new Date().toISOString(),
        result: `Price set to $${testPrice}`
      });

      // Step 2: Immediately check the price is locked
      testResults.steps.push({ step: 2, action: "Checking immediate lock status", timestamp: new Date().toISOString() });
      
      const immediateCheck = await authenticatedFetch(`${API_URL}/stocks/${testStockId}`, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      });

      if (!immediateCheck.ok) {
        throw new Error("Failed to fetch stock for immediate check");
      }

      const immediateData = await immediateCheck.json();
      const immediatePrice = immediateData.current_price;

      console.log(`🔒 Step 2 complete: Immediate price check = $${immediatePrice}`);
      testResults.steps.push({ 
        step: 2, 
        action: "Immediate price check completed", 
        timestamp: new Date().toISOString(),
        result: `Price is $${immediatePrice} (expected $${testPrice})`
      });

      if (Math.abs(immediatePrice - testPrice) > 0.01) {
        throw new Error(`Price not locked! Expected $${testPrice}, got $${immediatePrice}`);
      }

      // Step 3: Wait 5 seconds and check again
      testResults.steps.push({ step: 3, action: "Waiting 5 seconds...", timestamp: new Date().toISOString() });
      await new Promise(resolve => setTimeout(resolve, 5000));

      const fiveSecCheck = await authenticatedFetch(`${API_URL}/stocks/${testStockId}`, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      });

      if (!fiveSecCheck.ok) {
        throw new Error("Failed to fetch stock for 5-second check");
      }

      const fiveSecData = await fiveSecCheck.json();
      const fiveSecPrice = fiveSecData.current_price;

      console.log(`🔒 Step 3 complete: 5-second price check = $${fiveSecPrice}`);
      testResults.steps.push({ 
        step: 3, 
        action: "5-second price check completed", 
        timestamp: new Date().toISOString(),
        result: `Price is $${fiveSecPrice} (should still be locked at $${testPrice})`
      });

      if (Math.abs(fiveSecPrice - testPrice) > 0.01) {
        throw new Error(`Price changed too early! Expected $${testPrice}, got $${fiveSecPrice}`);
      }

      // Step 4: Wait another 10 seconds (total 15) and check again
      testResults.steps.push({ step: 4, action: "Waiting another 10 seconds (total 15s)...", timestamp: new Date().toISOString() });
      await new Promise(resolve => setTimeout(resolve, 10000));

      const fifteenSecCheck = await authenticatedFetch(`${API_URL}/stocks/${testStockId}`, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      });

      if (!fifteenSecCheck.ok) {
        throw new Error("Failed to fetch stock for 15-second check");
      }

      const fifteenSecData = await fifteenSecCheck.json();
      const fifteenSecPrice = fifteenSecData.current_price;

      console.log(`🔒 Step 4 complete: 15-second price check = $${fifteenSecPrice}`);
      testResults.steps.push({ 
        step: 4, 
        action: "15-second price check completed", 
        timestamp: new Date().toISOString(),
        result: `Price is $${fifteenSecPrice} (should still be locked at $${testPrice})`
      });

      if (Math.abs(fifteenSecPrice - testPrice) > 0.01) {
        throw new Error(`Price changed too early! Expected $${testPrice}, got $${fifteenSecPrice}`);
      }

      // Step 5: Wait another 20 seconds (total 35) to verify lock expires
      testResults.steps.push({ step: 5, action: "Waiting another 20 seconds (total 35s) for lock to expire...", timestamp: new Date().toISOString() });
      await new Promise(resolve => setTimeout(resolve, 20000));

      const finalCheck = await authenticatedFetch(`${API_URL}/stocks/${testStockId}`, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      });

      if (!finalCheck.ok) {
        throw new Error("Failed to fetch stock for final check");
      }

      const finalData = await finalCheck.json();
      const finalPrice = finalData.current_price;

      console.log(`🔒 Step 5 complete: Final price check = $${finalPrice}`);
      testResults.steps.push({ 
        step: 5, 
        action: "Final price check completed", 
        timestamp: new Date().toISOString(),
        result: `Price is $${finalPrice} (lock should be expired, price may have changed)`
      });

      // Step 6: Wait 5 more seconds to ensure price doesn't snap back to original
      testResults.steps.push({ step: 6, action: "Waiting 5 seconds post-lock to check for snapback...", timestamp: new Date().toISOString() });
      await new Promise(resolve => setTimeout(resolve, 5000));

      const postLockCheck = await authenticatedFetch(`${API_URL}/stocks/${testStockId}`, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      });

      if (!postLockCheck.ok) {
        throw new Error("Failed to fetch stock for post-lock check");
      }

      const postLockData = await postLockCheck.json();
      const postLockPrice = postLockData.current_price;

      console.log(`🔒 Step 6 complete: Post-lock price check = $${postLockPrice}`);
      testResults.steps.push({ 
        step: 6, 
        action: "Post-lock snapback check completed", 
        timestamp: new Date().toISOString(),
        result: `Price is $${postLockPrice} (checking for snapback to original $${originalPrice})`
      });

      // Check if price snapped back to original (which would be bad)
      const priceDifferenceFromOriginal = Math.abs(postLockPrice - originalPrice);
      const priceDifferenceFromTest = Math.abs(postLockPrice - testPrice);
      
      if (priceDifferenceFromOriginal < 0.01 && priceDifferenceFromTest > 0.02) {
        throw new Error(`Price snapped back! Post-lock price $${postLockPrice} is too close to original $${originalPrice}, suggesting snapback occurred`);
      }

      // Success if we made it this far
      testResults.success = true;
      testResults.steps.push({ 
        step: 7, 
        action: "Test completed successfully", 
        timestamp: new Date().toISOString(),
        result: `Stock price locking mechanism working correctly. No snapback detected. Final price: $${postLockPrice}`
      });

      console.log("🔒 Stock price locking test completed successfully!");
      return testResults;

    } catch (stepError) {
      testResults.error = stepError.message;
      testResults.steps.push({ 
        step: "error", 
        action: "Test failed", 
        timestamp: new Date().toISOString(),
        result: stepError.message
      });
      throw stepError;
    }

  } catch (error) {
    console.error("🔒 Error in stock price locking test:", error);
    throw error;
  }
};
