// auth/auth.ts
class AuthService {
  private static readonly ACCESS_TOKEN_KEY = 'access_token';
  private static readonly BACKEND_URL = import.meta.env.VITE_BACKEND_URL;

  private static isLoginPage(): boolean {
    return window.location.pathname === '/login';
  }

  static getUserRole(): 'admin' | 'viewer' | null {
    try {
      const token = this.getAccessToken();
      if (!token) return null;

      const payload = JSON.parse(atob(token.split('.')[1]));
      return payload.role || null;
    } catch (error) {
      console.error('Error parsing token for role:', error);
      return null;
    }
  }

  static getUserInfo(): { username?: string; role?: 'admin' | 'viewer' } | null {
    try {
      const token = this.getAccessToken();
      if (!token) return null;

      const payload = JSON.parse(atob(token.split('.')[1]));
      return {
        username: payload.username || payload.sub,
        role: payload.role,
      };
    } catch (error) {
      console.error('Error parsing token for user info:', error);
      return null;
    }
  }

  static isAdmin(): boolean {
    return this.getUserRole() === 'admin';
  }

  static isViewer(): boolean {
    return this.getUserRole() === 'viewer';
  }

  static canAccessRoute(path: string): boolean {
    const role = this.getUserRole();
    if (!role) return false;

    if (role === 'admin') return true;

    const viewerAllowedRoutes = ['/health', '/alert', '/log'];
    return viewerAllowedRoutes.includes(path);
  }

  static async authorized(): Promise<boolean> {
    try {
      if (this.isLoginPage()) {
        return false;
      }

      const accessToken = this.getAccessToken();
      if (!accessToken) {
        console.log('No access token found, attempting to refresh...');
        return await this.attemptTokenRefresh();
      }

      if (this.isTokenExpired(accessToken)) {
        console.log('Access token expired, attempting to refresh...');
        return await this.attemptTokenRefresh();
      }
      return true;
    } catch (error) {
      console.error('Authorization check failed:', error);
      return false;
    }
  }

  private static async attemptTokenRefresh(): Promise<boolean> {
    try {
      const isRefreshTokenValid = await this.verifyRefreshToken();

      if (!isRefreshTokenValid) {
        console.log('Refresh token is invalid or missing.');
        this.clearTokens();
        if (!this.isLoginPage()) this.redirectToLogin();
        return false;
      }

      const response = await fetch(`${this.BACKEND_URL}/api/auth/refresh`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
      });

      if (response.ok) {
        const text = await response.text();
        if (text) {
          const data = JSON.parse(text);
          if (data.status === 'ok' && data.access_token) {
            this.setAccessToken(data.access_token);
            console.log('Token refreshed successfully');
            return true;
          }
        }
      }

      console.log('Token refresh failed');
      this.clearTokens();
      if (!this.isLoginPage()) this.redirectToLogin();
      return false;

    } catch (error) {
      console.error('Token refresh error:', error);
      this.clearTokens();
      if (!this.isLoginPage()) this.redirectToLogin();
      return false;
    }
  }

  private static async verifyRefreshToken(): Promise<boolean> {
    try {
      const response = await fetch(`${this.BACKEND_URL}/api/auth/verify`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
      });

      if (response.ok) {
        const text = await response.text();
        if (text) {
          const data = JSON.parse(text);
          return data.status === 'authorized';
        }
      }

      return false;
    } catch (error) {
      console.error('Error verifying refresh token:', error);
      return false;
    }
  }

  static getAccessToken(): string | null {
    return localStorage.getItem(this.ACCESS_TOKEN_KEY);
  }

  static setAccessToken(token: string): void {
    localStorage.setItem(this.ACCESS_TOKEN_KEY, token);
  }

  static clearTokens(): void {
    localStorage.removeItem(this.ACCESS_TOKEN_KEY);
  }

  private static isTokenExpired(token: string): boolean {
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      const currentTime = Math.floor(Date.now() / 1000);
      return payload.exp < currentTime;
    } catch (error) {
      console.error('Error parsing token:', error);
      return true;
    }
  }

  public static redirectToLogin(): void {
    if (!this.isLoginPage()) {
      window.location.href = '/login';
    }
  }

  static getAuthHeader(): { Authorization: string } | Record<string, never> {
    const token = this.getAccessToken();
    return token ? { Authorization: `Bearer ${token}` } : {};
  }

  static getApiHeaders(additionalHeaders: Record<string, string> = {}): Record<string, string> {
    const authHeaders = this.getAuthHeader();
    return {
      'Content-Type': 'application/json',
      ...authHeaders,
      ...additionalHeaders,
    };
  }

  static async makeAuthenticatedRequest(
    url: string,
    options: RequestInit = {},
    retry = true
  ): Promise<Response> {
    const isAuthorized = await this.authorized();
    if (!isAuthorized) {
      if (!this.isLoginPage()) this.redirectToLogin();
      throw new Error('Not authorized');
    }

    const authHeaders = this.getAuthHeader();
    const headers = {
      'Content-Type': 'application/json',
      ...authHeaders,
      ...(options.headers || {}),
    };

    const response = await fetch(url, {
      ...options,
      headers,
      credentials: 'include',
    });

    if (response.status === 401 && retry) {
      console.log('401 received, trying token refresh...');
      const refreshed = await this.attemptTokenRefresh();

      if (refreshed) {
        const newAuthHeaders = this.getAuthHeader();
        const retryHeaders = {
          'Content-Type': 'application/json',
          ...newAuthHeaders,
          ...(options.headers || {}),
        };

        return fetch(url, {
          ...options,
          headers: retryHeaders,
          credentials: 'include',
        });
      } else {
        if (!this.isLoginPage()) this.redirectToLogin();
        throw new Error('Authentication failed');
      }
    }

    return response;
  }

  static async logout(): Promise<void> {
    this.clearTokens();
    if (!this.isLoginPage()) this.redirectToLogin();
  }
}

export default AuthService;
