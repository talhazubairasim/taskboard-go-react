export interface RegistrationForm {
  name: string;
  email: string;
  password: string;
  confirmPassword: string;
}

/** Mirrors the server's password rules so users get instant feedback. */
export function validateRegistration(f: RegistrationForm): Record<string, string> {
  const errors: Record<string, string> = {};
  if (!f.name.trim()) errors.name = 'Name is required';
  if (!f.email.trim()) errors.email = 'Email is required';
  else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(f.email)) errors.email = 'Invalid email format';

  if (!f.password) errors.password = 'Password is required';
  else if (f.password.length < 8) errors.password = 'Password must be at least 8 characters';
  else if (!/[a-z]/.test(f.password)) errors.password = 'Password must contain a lowercase letter';
  else if (!/[A-Z]/.test(f.password)) errors.password = 'Password must contain an uppercase letter';
  else if (!/\d/.test(f.password)) errors.password = 'Password must contain a number';
  else if (!/[^A-Za-z0-9\s]/.test(f.password)) errors.password = 'Password must contain a special character';

  if (f.password !== f.confirmPassword) errors.confirmPassword = 'Passwords do not match';
  return errors;
}
