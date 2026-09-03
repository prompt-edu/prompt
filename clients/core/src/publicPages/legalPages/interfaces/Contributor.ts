export interface Contributor {
  login: string
  avatar_url: string
  html_url: string
  contributions: number
  type: string
}

export interface ContributorWithInfo extends Contributor {
  name: string
  contribution: string
  pinnedPosition?: number
}
