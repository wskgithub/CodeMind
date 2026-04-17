import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import LoginPage from '../login/LoginPage';

// mock functions must be defined before vi.mock
const mockFns = {
  messageSuccess: vi.fn(),
  messageError: vi.fn(),
  login: vi.fn(),
  navigate: vi.fn(),
};

// Mock antd message
vi.mock('antd', async () => {
  const actual = await vi.importActual('antd');
  return {
    ...(actual as object),
    message: {
      success: (...args: unknown[]) => mockFns.messageSuccess(...args),
      error: (...args: unknown[]) => mockFns.messageError(...args),
    },
  };
});

// mock auth store with selector pattern
vi.mock('@/store/authStore', () => ({
  default: (selector: (state: { login: typeof mockFns.login }) => typeof mockFns.login) => {
    const store = { login: mockFns.login };
    return selector ? selector(store) : store;
  },
}));

// mock app store with selector pattern
vi.mock('@/store/appStore', () => ({
  default: (selector?: (state: Record<string, unknown>) => unknown) => {
    const store = {
      themeMode: 'dark',
      toggleTheme: vi.fn(),
      setLanguage: vi.fn(),
      language: 'zh-CN',
    };
    return selector ? selector(store) : store;
  },
}));

// Mock useNavigate and useLocation
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom');
  return {
    ...(actual as object),
    useNavigate: () => mockFns.navigate,
    useLocation: () => ({ state: { from: { pathname: '/dashboard' } } }),
  };
});

// Mock axios
vi.mock('axios', () => ({
  default: {
    isAxiosError: (error: unknown) => error && typeof error === 'object' && 'response' in error,
  },
}));

describe('LoginPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should render login form correctly', () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    );

    expect(screen.getByPlaceholderText('用户名')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('密码')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /登 录/i })).toBeInTheDocument();
    expect(screen.getByText('CodeMind')).toBeInTheDocument();
  });

  it('should show validation error when submitting empty form', async () => {
    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    );

    const submitButton = screen.getByRole('button', { name: /登 录/i });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText('请输入用户名')).toBeInTheDocument();
      expect(screen.getByText('请输入密码')).toBeInTheDocument();
    });
  });

  it('should handle successful login', async () => {
    mockFns.login.mockResolvedValueOnce(undefined);

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    );

    const usernameInput = screen.getByPlaceholderText('用户名');
    const passwordInput = screen.getByPlaceholderText('密码');
    const submitButton = screen.getByRole('button', { name: /登 录/i });

    fireEvent.change(usernameInput, { target: { value: 'admin' } });
    fireEvent.change(passwordInput, { target: { value: 'Admin@123456' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(mockFns.login).toHaveBeenCalledWith('admin', 'Admin@123456');
      expect(mockFns.messageSuccess).toHaveBeenCalledWith('登录成功');
      expect(mockFns.navigate).toHaveBeenCalledWith('/dashboard', { replace: true });
    });
  });

  it('should handle login error with 401 status', async () => {
    const error = {
      response: {
        status: 401,
        data: {
          code: 401,
          message: '用户名或密码错误',
          data: { fail_count: 1, max_fail_count: 5 },
        },
      },
    };
    mockFns.login.mockRejectedValueOnce(error);

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    );

    const usernameInput = screen.getByPlaceholderText('用户名');
    const passwordInput = screen.getByPlaceholderText('密码');
    const submitButton = screen.getByRole('button', { name: /登 录/i });

    fireEvent.change(usernameInput, { target: { value: 'admin' } });
    fireEvent.change(passwordInput, { target: { value: 'wrongpassword' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(mockFns.messageError).toHaveBeenCalledWith(expect.stringContaining('用户名或密码错误'));
    });
  });

  it('should handle account lock error', async () => {
    const error = {
      response: {
        status: 403,
        data: {
          code: 40008,
          message: '账号已被锁定',
          data: {
            locked: true,
            remaining_time: 300,
            fail_count: 5,
            max_fail_count: 5,
          },
        },
      },
    };
    mockFns.login.mockRejectedValueOnce(error);

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    );

    const usernameInput = screen.getByPlaceholderText('用户名');
    const passwordInput = screen.getByPlaceholderText('密码');
    const submitButton = screen.getByRole('button', { name: /登 录/i });

    fireEvent.change(usernameInput, { target: { value: 'admin' } });
    fireEvent.change(passwordInput, { target: { value: 'wrongpassword' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText('账号已被锁定')).toBeInTheDocument();
      expect(screen.getByText(/剩余时间/)).toBeInTheDocument();
    });
  });

  it('should show loading state during login', async () => {
    mockFns.login.mockImplementation(() => new Promise(() => {}));

    const { container } = render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    );

    const usernameInput = screen.getByPlaceholderText('用户名');
    const passwordInput = screen.getByPlaceholderText('密码');
    const submitButton = screen.getByRole('button', { name: /登 录/i });

    fireEvent.change(usernameInput, { target: { value: 'admin' } });
    fireEvent.change(passwordInput, { target: { value: 'password' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      // Check for loading spinner in the button
      const loadingIcon = container.querySelector('.ant-btn-loading-icon');
      expect(loadingIcon).toBeInTheDocument();
    });
  });

  it('should disable form when account is locked', async () => {
    const error = {
      response: {
        status: 403,
        data: {
          code: 40008,
          message: '账号已被锁定',
          data: {
            locked: true,
            remaining_time: 300,
            fail_count: 5,
            max_fail_count: 5,
          },
        },
      },
    };
    mockFns.login.mockRejectedValueOnce(error);

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>
    );

    const usernameInput = screen.getByPlaceholderText('用户名');
    const passwordInput = screen.getByPlaceholderText('密码');
    const submitButton = screen.getByRole('button', { name: /登 录/i });

    fireEvent.change(usernameInput, { target: { value: 'admin' } });
    fireEvent.change(passwordInput, { target: { value: 'wrongpassword' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByPlaceholderText('用户名')).toBeDisabled();
      expect(screen.getByPlaceholderText('密码')).toBeDisabled();
      expect(screen.getByRole('button', { name: /账号已锁定/i })).toBeDisabled();
    });
  });
});
