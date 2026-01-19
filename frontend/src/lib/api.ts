// API client for communicating with the Go backend
const API_BASE = "/api";

export interface Book {
  id: number;
  book_id: number;
  title: string;
  description: string;
  series_name?: string;
  series_number?: number;
  series_id?: number;
  asin?: string;
  isbn10?: string;
  isbn13?: string;
  language?: string;
  hardcover_id?: string;
  hardcover_book_id?: number;
  goodreads_id?: string;
  google_id?: string;
  authors?: string[];
  is_missing?: boolean;
}

export interface Series {
  id: number;
  series_id: number;
  name: string;
  description?: string;
  url?: string;
  authors?: string[];
}

export interface SyncSeriesResponse {
  status: string;
  message: string;
  series_id: number;
  existing_books: number;
  missing_books: number;
  new_missing_books: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  per_page: number;
}

async function fetchApi<T>(
  endpoint: string,
  options?: RequestInit,
): Promise<T> {
  const response = await fetch(`${API_BASE}${endpoint}`, {
    headers: {
      "Content-Type": "application/json",
      ...options?.headers,
    },
    ...options,
  });

  if (!response.ok) {
    console.error(`API error: ${response.status} ${response.statusText}`);
    throw new Error(`API error: ${response.status} ${response.statusText}`);
  }

  return response.json();
}

export const api = {
  // Books
  async getBooks(page = 1, perPage = 20): Promise<PaginatedResponse<Book>> {
    return fetchApi<PaginatedResponse<Book>>(
      `/books?page=${page}&per_page=${perPage}`,
    );
  },

  // Series
  async getSeries(page = 1, perPage = 20): Promise<PaginatedResponse<Series>> {
    return fetchApi<PaginatedResponse<Series>>(
      `/series?page=${page}&per_page=${perPage}`,
    );
  },

  async getSeriesById(id: number): Promise<Series> {
    return fetchApi<Series>(`/series/${id}`);
  },

  async getSeriesBooks(id: number): Promise<Book[]> {
    return fetchApi<Book[]>(`/series/${id}/books`);
  },

  async fetchSeriesFromGoodreads(id: number): Promise<SyncSeriesResponse> {
    return fetchApi<SyncSeriesResponse>(`/series/${id}/goodreads`, {
      method: "POST",
    });
  },

  // Config
  async getConfig(): Promise<{
    serverUrl: string;
    username: string;
    password: string;
  }> {
    console.log("Fetching config from API");
    return fetchApi<{ serverUrl: string; username: string; password: string }>(
      "/config",
    );
  },

  async saveConfig(
    serverUrl: string,
    username: string,
    password: string,
  ): Promise<void> {
    return fetchApi<void>("/config", {
      method: "POST",
      body: JSON.stringify({ serverUrl, username, password }),
    });
  },
};
