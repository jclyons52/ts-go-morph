import { helper } from "./helper";

export type ID = string;

export enum Status {
  Active,
  Inactive
}

export const DEFAULT_ID: ID = "none";

export function isActive(status: Status): boolean {
  return status === Status.Active;
}
