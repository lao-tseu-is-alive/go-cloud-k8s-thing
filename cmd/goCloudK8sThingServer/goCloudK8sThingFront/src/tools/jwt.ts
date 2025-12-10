// Represents a numeric date (Unix timestamp)
type NumericDate = number | null;

// Audience can be a string or array of strings
type Audience = string | string[];

// RegisteredClaims structure
export interface RegisteredClaims {
  jti?: string;
  aud?: Audience;
  iss?: string;
  sub?: string;
  exp?: NumericDate;
  iat?: NumericDate;
  nbf?: NumericDate;
}

// UserInfo structure for custom claims
export interface UserInfo {
  user_id: number;
  external_id: number;
  name: string;
  email: string;
  login: string;
  is_admin: boolean;
  groups: number[];
}

// JwtCustomClaims structure combining RegisteredClaims and UserInfo
export interface JwtCustomClaims extends RegisteredClaims {
  User?: UserInfo;
}
