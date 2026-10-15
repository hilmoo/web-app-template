import { ConnectError } from "@connectrpc/connect";

const FALLBACK = "Something went wrong";

type Detail = { type?: string; debug?: unknown };

function isRecord(v: unknown): v is Record<string, unknown> {
  return typeof v === "object" && v !== null;
}

function getDetails(error: unknown): Detail[] {
  // ConnectError instance (what the client normally throws)
  if (error instanceof ConnectError) {
    return error.details as Detail[];
  }
  // Plain object shaped like the Connect JSON error
  if (isRecord(error) && Array.isArray(error.details)) {
    return error.details as Detail[];
  }
  return [];
}

function findDebug(details: Detail[], type: string): Record<string, unknown> | undefined {
  const d = details.find((d) => d.type === type);
  return isRecord(d?.debug) ? d.debug : undefined;
}

export function getErrorMessage(error: unknown): string {
  const details = getDetails(error);

  // 1. Validation errors: "name: This field is required"
  const badRequest = findDebug(details, "google.rpc.BadRequest");
  const violations = badRequest?.fieldViolations;
  if (Array.isArray(violations) && violations.length > 0) {
    const messages = violations
      .filter(isRecord)
      .map((v) => {
        const field = typeof v.field === "string" ? v.field : "";
        const desc = typeof v.description === "string" ? v.description : "";
        return field && desc ? `${field}: ${desc}` : desc || field;
      })
      .filter(Boolean);
    if (messages.length > 0) {
      return messages.join("\n");
    }
  }

  // 2. Specific reason, e.g. "invalid email or password"
  const errorInfo = findDebug(details, "google.rpc.ErrorInfo");
  if (typeof errorInfo?.reason === "string" && errorInfo.reason) {
    return errorInfo.reason;
  }

  // 3. Generic message
  if (error instanceof ConnectError) {
    return error.rawMessage || FALLBACK;
  }
  if (isRecord(error) && typeof error.message === "string" && error.message) {
    return error.message;
  }
  return FALLBACK;
}
