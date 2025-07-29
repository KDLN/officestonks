import { authenticatedFetch } from "./authBridge";

const isLocalhost =
  window.location.hostname === "localhost" ||
  window.location.hostname === "127.0.0.1";
const BACKEND_URL =
  process.env.REACT_APP_BACKEND_URL ||
  "https://web-production-1e26.up.railway.app";
const BASE_URL = isLocalhost ? "/api" : `${BACKEND_URL}/api`;
const API_URL = BASE_URL.endsWith("/api") ? BASE_URL : `${BASE_URL}/api`;

export const fetchActiveNews = async () => {
  const response = await authenticatedFetch(`${API_URL}/news`, {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
  });
  if (!response.ok) {
    throw new Error("Failed to fetch news");
  }
  return response.json();
};
