// Keycloak caps usernames and attribute values at 255 characters
export const MAX_UNIVERSITY_LOGIN_LENGTH = 255

// The server looks the login up as the Keycloak attribute query `university_login:<login>`, which
// whitespace and colons would break
const FORBIDDEN_CHARACTERS = /[\s\p{Cc}:]/u

export const validateUniversityLogin = (universityLogin: string): string | undefined => {
  if (universityLogin === '') {
    return 'A university login is required.'
  }
  if ([...universityLogin].length > MAX_UNIVERSITY_LOGIN_LENGTH) {
    return `The university login must be at most ${MAX_UNIVERSITY_LOGIN_LENGTH} characters.`
  }
  if (FORBIDDEN_CHARACTERS.test(universityLogin)) {
    return 'The university login cannot contain spaces or colons.'
  }
  return undefined
}
