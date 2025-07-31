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
