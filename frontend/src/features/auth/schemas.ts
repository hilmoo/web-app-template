import * as z from "zod";
import { AuthLoginBody, AuthRegisterBody, AuthUpdateMeBody } from "#/api/schemas/auth/auth";

const requiredEmail = z.string().min(1, "Email is required").pipe(z.email("Enter a valid email"));
const requiredPassword = z.string().min(1, "Password is required");
const requiredFullName = z.string().min(1, "Full name is required");

export const loginSchema = AuthLoginBody.extend({
  email: requiredEmail,
  password: requiredPassword,
});
export type loginSchemaTypes = z.input<typeof loginSchema>;

export const registerSchema = AuthRegisterBody.extend({
  email: requiredEmail,
  fullName: requiredFullName,
  password: requiredPassword,
});
export type registerSchemaTypes = z.input<typeof registerSchema>;

export const updateMeSchema = AuthUpdateMeBody.extend({
  email: requiredEmail,
  fullName: requiredFullName,
});
export type updateMeSchemaTypes = z.input<typeof updateMeSchema>;
