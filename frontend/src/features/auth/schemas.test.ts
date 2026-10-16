import { describe, expect, it } from "vitest";
import { loginSchema, registerSchema, updateMeSchema } from "./schemas";

describe("loginSchema", () => {
  it("accepts a valid email and password", () => {
    expect(loginSchema.safeParse({ email: "jane@example.com", password: "pw" }).success).toBe(true);
  });

  it("rejects an empty email with a friendly message", () => {
    const result = loginSchema.safeParse({ email: "", password: "pw" });
    expect(result.success).toBe(false);
    expect(result.error?.issues[0]?.message).toBe("Email is required");
  });

  it("rejects a malformed email", () => {
    expect(loginSchema.safeParse({ email: "not-an-email", password: "pw" }).success).toBe(false);
  });
});

describe("registerSchema", () => {
  it("requires a full name", () => {
    const result = registerSchema.safeParse({
      email: "jane@example.com",
      fullName: "",
      password: "pw",
    });
    expect(result.success).toBe(false);
  });
});

describe("updateMeSchema", () => {
  it("accepts a valid profile", () => {
    expect(
      updateMeSchema.safeParse({ email: "jane@example.com", fullName: "Jane Doe" }).success,
    ).toBe(true);
  });
});
