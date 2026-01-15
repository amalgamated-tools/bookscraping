// API client for communicating with the Go backend
const API_BASE = '/api';

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
}

export interface Series {
    id: number;
    series_id: number;
    name: string;
    description?: string;
    url?: string;
}

export interface PaginatedResponse<T> {
    data: T[];
    total: number;
    page: number;
    per_page: number;
}

export interface BookloreMetadata {
    bookId: number;
    title: string;
    description?: string;
    publisher?: string;
    publishedDate?: string;
    seriesName?: string;
    seriesNumber?: number;
    seriesTotal?: number;
    isbn13?: string;
    isbn10?: string;
    asin?: string;
    pageCount?: number;
    language?: string;
    goodreadsId?: string;
    googleId?: string;
    hardcoverId?: string;
    hardcoverBookId?: number;
    authors?: string[];
    categories?: string[];
}

export interface BookloreBook {
    id: number;
    bookType: string;
    libraryId: number;
    fileName: string;
    addedOn: string;
    metadata: BookloreMetadata;
}

export interface BookloreAuthResponse {
    isDefaultPassword: string;
    accessToken: string;
    refreshToken: string;
}

async function fetchApi<T>(endpoint: string, options?: RequestInit): Promise<T> {
    const response = await fetch(`${API_BASE}${endpoint}`, {
        headers: {
            'Content-Type': 'application/json',
            ...options?.headers
        },
        ...options
    });

    if (!response.ok) {
        throw new Error(`API error: ${response.status} ${response.statusText}`);
    }

    return response.json();
}

export const api = {
    // Books
    async getBooks(page = 1, perPage = 20): Promise<PaginatedResponse<Book>> {
        return fetchApi<PaginatedResponse<Book>>(`/books?page=${page}&per_page=${perPage}`);
    },

    async getBook(id: number): Promise<Book> {
        return fetchApi<Book>(`/books/${id}`);
    },

    async searchBooks(query: string): Promise<Book[]> {
        return fetchApi<Book[]>(`/books/search?q=${encodeURIComponent(query)}`);
    },

    // Series
    async getSeries(page = 1, perPage = 20): Promise<PaginatedResponse<Series>> {
        return fetchApi<PaginatedResponse<Series>>(`/series?page=${page}&per_page=${perPage}`);
    },

    async getSeriesById(id: number): Promise<Series> {
        return fetchApi<Series>(`/series/${id}`);
    },

    // Goodreads integration
    async searchGoodreads(query: string): Promise<Book[]> {
        return fetchApi<Book[]>(`/goodreads/search?q=${encodeURIComponent(query)}`);
    },

    async getGoodreadsBook(id: string): Promise<Book> {
        return fetchApi<Book>(`/goodreads/book/${id}`);
    },

    async getGoodreadsSeries(id: string): Promise<Series> {
        return fetchApi<Series>(`/goodreads/series/${id}`);
    },

    // Booklore authentication
    async bookloreLogin(serverUrl: string, username: string, password: string): Promise<BookloreAuthResponse> {
        const response = await fetch(`${serverUrl}/api/v1/auth/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ username, password })
        });

        if (!response.ok) {
            throw new Error(`Booklore login failed: ${response.status} ${response.statusText}`);
        }

        return response.json();
    },

    // Booklore API proxy
    async getBookloreBooks(serverUrl: string, token: string): Promise<BookloreBook[]> {
        const response = await fetch(`${serverUrl}/api/v1/books`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (!response.ok) {
            throw new Error(`Failed to fetch books from Booklore: ${response.status} ${response.statusText}`);
        }

        return response.json();
    },

    async syncBooks(server_url?: string, username?: string, password?: string): Promise<void> {
        return fetchApi<void>('/sync', {
            method: 'POST',
            body: JSON.stringify({ server_url, username, password })
        });
    },

    // Config
    async getConfig(): Promise<{ serverUrl: string, username: string, password: string }> {
        return fetchApi<{ serverUrl: string, username: string, password: string }>('/config');
    },

    async saveConfig(serverUrl: string, username: string, password: string): Promise<void> {
        return fetchApi<void>('/config', {
            method: 'POST',
            body: JSON.stringify({ serverUrl, username, password })
        });
    }
};
