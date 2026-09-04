# Errors

- **Errors wrap with context** (`fmt.Errorf("…: %w", err)`) and reach
  the user as [`product/rules.md`](../../product/rules.md) demands: the
  failing check named, plain text, never a stack trace.
