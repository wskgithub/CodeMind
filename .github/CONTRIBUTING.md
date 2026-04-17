# Contributing to CodeMind

Thank you for your interest in contributing to CodeMind! This document outlines the development guidelines and best practices for the project.

## Table of Contents

- [Code Standards](#code-standards)
- [Git Workflow](#git-workflow)
- [Commit Conventions](#commit-conventions)
- [Code Review](#code-review)
- [Testing Requirements](#testing-requirements)

---

## Code Standards

### Go Backend Standards

**Code Style**
- Follow the [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` for code formatting
- Use `golangci-lint` for static analysis

**Naming Conventions**
- Package names: lowercase words, no underscores or camelCase
- Interface names: use `-er` suffix for behavior-based interfaces
- Constants: camelCase, exported constants start with uppercase

**Documentation**
- All exported functions must have doc comments
- Package comments should be in a `doc.go` file
- Use godoc format

```go
// UserService handles user-related business logic
type UserService struct {
    repo repository.UserRepository
}

// CreateUser creates a new user account
// Parameters: ctx - request context, req - creation request
// Returns: created user info and error
func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*model.User, error) {
    // ...
}
```

**Error Handling**
- Never ignore errors
- Use `errors.Wrap` to add context to errors
- Define clear error codes and messages

### TypeScript Frontend Standards

**Code Style**
- Follow the [Airbnb TypeScript Style Guide](https://github.com/airbnb/typescript)
- Use ESLint + Prettier for formatting
- Use functional components with Hooks

**Naming Conventions**
- Components: PascalCase (e.g., `UserList.tsx`)
- Utilities: camelCase (e.g., `formatDate.ts`)
- Constants: UPPER_SNAKE_CASE
- Types/Interfaces: PascalCase (e.g., `UserData`)

**Component Structure**
```typescript
// ✅ Component file structure
// 1. Imports
import { useState } from 'react';
import type { User } from '@/types';

// 2. Type definitions
interface UserListProps {
  users: User[];
  onEdit: (user: User) => void;
}

// 3. Component definition
export function UserList({ users, onEdit }: UserListProps) {
  // 4. Hooks
  const [filter, setFilter] = useState('');

  // 5. Side effects
  useEffect(() => {
    // ...
  }, []);

  // 6. Event handlers
  const handleEdit = (user: User) => {
    onEdit(user);
  };

  // 7. Render
  return (
    <div className="user-list">
      {/* ... */}
    </div>
  );
}
```

---

## Git Workflow

### Branch Model

```
main (production)
  ↑
  ├── develop (development)
       ↑
       ├── feature/* (feature branches)
       ├── fix/* (bugfix branches)
       └── release/* (release branches)
```

### Branch Naming Convention

| Branch Type | Format | Example |
|-------------|--------|---------|
| Feature | `feature/<module>-<feature>` | `feature/user-management` |
| Bug Fix | `fix/<module>-<issue>` | `fix/auth-login-validation` |
| Release | `release/<version>` | `release/0.1.0` |
| Hotfix | `hotfix/<version>-<issue>` | `hotfix/1.0.0-security-fix` |

### Branch Workflow

1. **Create a feature branch from develop**
   ```bash
   git checkout develop
   git pull origin develop
   git checkout -b feature/user-management
   ```

2. **Develop and commit**
   ```bash
   git add .
   git commit -m "feat(user): add user creation API"
   ```

3. **Push to remote**
   ```bash
   git push origin feature/user-management
   ```

4. **Create a Pull Request**
   - Target branch: `develop`
   - Fill out the PR description template
   - Request code review

5. **Clean up after merge**
   ```bash
   git checkout develop
   git pull origin develop
   git branch -d feature/user-management
   ```

---

## Commit Conventions

### Conventional Commits

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Commit Types

| Type | Description | Example |
|------|-------------|---------|
| `feat` | New feature | `feat(user): add user creation API` |
| `fix` | Bug fix | `fix(auth): resolve token expiration issue` |
| `docs` | Documentation changes | `docs(readme): update deployment instructions` |
| `style` | Code formatting (no logic changes) | `style(backend): format code with prettier` |
| `refactor` | Code refactoring | `refactor(service): extract common validation logic` |
| `perf` | Performance improvements | `perf(cache): add redis caching for user data` |
| `test` | Testing related | `test(auth): add login unit tests` |
| `chore` | Build/tooling changes | `chore(deps): upgrade gin framework to v1.10.0` |

### Scope

**Backend**: `auth`, `user`, `department`, `apikey`, `stats`, `limit`, `system`, `llm`, `db`
**Frontend**: `login`, `dashboard`, `admin`, `components`, `api`, `style`
**Deployment**: `docker`, `deploy`, `ci`

### Commit Examples

```bash
# Feature
git commit -m "feat(user): add batch import users from CSV

- Parse CSV file with validation
- Create users in batch with transaction
- Return import summary with success/failure count

Closes #123"

# Bug fix
git commit -m "fix(auth): fix JWT blacklist not working

Use JTI instead of full token for blacklist key to avoid
URL encoding issues.

Fixes #145"
```

---

## Code Review

### PR Review Checklist

**Functionality**
- [ ] Feature matches requirements
- [ ] Edge cases are handled
- [ ] Error handling is complete

**Code Quality**
- [ ] Code style follows conventions
- [ ] Naming is clear and descriptive
- [ ] No duplicate code
- [ ] Comments are appropriate and accurate

**Test Coverage**
- [ ] Unit test coverage > 80%
- [ ] Integration tests for critical paths
- [ ] Test cases are comprehensive

**Security**
- [ ] Input validation is complete
- [ ] Sensitive data is encrypted
- [ ] SQL/command injection prevention
- [ ] Permission checks are correct

### Review Feedback Guidelines

**Constructive Feedback**
- Explain the reason when pointing out issues
- Provide improvement suggestions
- For style issues, directly fix or reference the guideline

**Labels**
- `LGTM` (Looks Good To Me): Review approved
- `Request Changes`: Changes required before approval
- `Concept ACK`: Design approved, details pending

---

## Testing Requirements

### Backend Testing

**Unit Tests**
- Service layer coverage > 80%
- Repository layer uses mock data
- Each function: at least one success path + one failure path

```go
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name    string
        req     *dto.CreateUserRequest
        want    *model.User
        wantErr bool
    }{
        {
            name: "create user successfully",
            req:  &dto.CreateUserRequest{Username: "test", ...},
            want: &model.User{ID: 1, Username: "test", ...},
        },
        {
            name:    "duplicate username should fail",
            req:     &dto.CreateUserRequest{Username: "existing", ...},
            wantErr: true,
        },
    }
    // ...
}
```

### Frontend Testing

**Component Tests**
- Use React Testing Library
- Test user interactions, not implementation details
- Mock API calls

```typescript
describe('LoginForm', () => {
  it('should show error on failed login', async () => {
    const mockLogin = vi.fn().mockRejectedValue(new Error('Invalid credentials'));
    render(<LoginForm onLogin={mockLogin} />);

    await userEvent.click(screen.getByRole('button', { name: /login/i }));

    expect(await screen.findByText('Invalid username or password')).toBeInTheDocument();
  });
});
```

---

## Security Guidelines

### Password Security
- Use bcrypt with cost factor = 12
- Minimum 8 characters with uppercase, lowercase, and numbers
- Never log passwords

### API Key Security
- Display full key only at creation time
- Store SHA-256 hash in database
- Log only key_prefix

### JWT Security
- Use HS256 algorithm for signing
- Token expiration: 24 hours
- Add to blacklist on logout

### Input Validation
- Validate all user input
- Use allowlist instead of blocklist
- Use parameterized SQL queries

---

## Resources

- [Development Standards](../docs/development-standards.md)
- [Backend Standards](../docs/backend-standards.md)
- [Frontend Standards](../docs/frontend-standards.md)
- [Testing Guide](../docs/testing-guide.md)
- [Architecture Overview](../docs/architecture.md)
