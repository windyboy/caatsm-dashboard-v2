import { browser } from '$app/environment';

const TOKEN_KEY = 'caatsm_auth_token';
const CSRF_KEY = 'caatsm_csrf_token';
const API_BASE = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:3002";

export interface AuthTokens {
  accessToken: string;
  csrfToken: string;
}

/**
 * 认证管理类
 */
export class AuthManager {
  private static instance: AuthManager;
  private tokens: AuthTokens | null = null;

  private constructor() {
    if (browser) {
      this.loadTokens();
    }
  }

  static getInstance(): AuthManager {
    if (!AuthManager.instance) {
      AuthManager.instance = new AuthManager();
    }
    return AuthManager.instance;
  }

  /**
   * 登录并存储token
   */
  async login(username: string, password: string): Promise<void> {
    const response = await fetch(`${API_BASE}/api/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username, password }),
    });

    if (!response.ok) {
      throw new Error(`Login failed: ${response.status}`);
    }

    const data = await response.json();
    this.tokens = {
      accessToken: data.access_token,
      csrfToken: data.csrf_token,
    };

    if (browser) {
      try {
        localStorage.setItem(TOKEN_KEY, this.tokens.accessToken);
        localStorage.setItem(CSRF_KEY, this.tokens.csrfToken);
      } catch (error) {
        // Handle localStorage errors (e.g., privacy mode, quota exceeded)
        console.warn('Failed to store tokens in localStorage:', error);
        // Tokens are still in memory, so authentication will work for this session
      }
    }
  }

  /**
   * 登出并清除token
   */
  logout(): void {
    this.tokens = null;
    if (browser) {
      try {
        localStorage.removeItem(TOKEN_KEY);
        localStorage.removeItem(CSRF_KEY);
      } catch (error) {
        // Handle localStorage errors gracefully
        console.warn('Failed to remove tokens from localStorage:', error);
      }
    }
  }

  /**
   * 获取访问token
   */
  getAccessToken(): string | null {
    return this.tokens?.accessToken || null;
  }

  /**
   * 获取CSRF token
   */
  getCsrfToken(): string | null {
    return this.tokens?.csrfToken || null;
  }

  /**
   * 检查是否已认证
   */
  isAuthenticated(): boolean {
    return !!this.getAccessToken();
  }

  /**
   * 从存储加载token
   */
  private loadTokens(): void {
    try {
      const accessToken = localStorage.getItem(TOKEN_KEY);
      const csrfToken = localStorage.getItem(CSRF_KEY);

      if (accessToken && csrfToken) {
        this.tokens = { accessToken, csrfToken };
      }
    } catch (error) {
      // Handle localStorage errors (e.g., privacy mode, disabled storage)
      console.warn('Failed to load tokens from localStorage:', error);
      // Continue without tokens - user will need to log in again
    }
  }

  /**
   * 刷新token
   */
  async refreshToken(): Promise<void> {
    if (!this.tokens?.accessToken) {
      throw new Error('No token to refresh');
    }

    const response = await fetch(`${API_BASE}/api/auth/refresh`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${this.tokens.accessToken}`,
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      this.logout();
      throw new Error('Token refresh failed');
    }

    const data = await response.json();
    this.tokens.accessToken = data.access_token;

    if (browser) {
      try {
        localStorage.setItem(TOKEN_KEY, this.tokens.accessToken);
      } catch (error) {
        // Handle localStorage errors gracefully
        console.warn('Failed to store refreshed token in localStorage:', error);
        // Token is still in memory, so authentication will work for this session
      }
    }
  }
}

export const authManager = AuthManager.getInstance();