export const contributorMapping: {
  [username: string]: {
    name: string
    contribution: string
    // Overrides the commit-count ordering. Only set this for entries whose place in the
    // list should not depend on how many commits they have.
    pinnedPosition?: number
  }
} = {
  Mtze: {
    name: 'Matthias Linhuber',
    contribution: 'Project Manager, Architect',
    pinnedPosition: 1,
  },
  airelawaleria: {
    name: 'Valeryia Andraichuk',
    contribution: 'Creator of PROMPT - the concept on which PROMPT 2.0 is based',
    pinnedPosition: 2,
  },
  niclasheun: {
    name: 'Niclas Heun',
    contribution: 'Lead Developer & Architect of PROMPT 2.0',
  },
  rappm: {
    name: 'Maximilian Rapp',
    contribution: 'Maintainer, Assessment Phase',
  },
  mathildeshagl: {
    name: 'Mathilde Hagl',
    contribution: 'Maintainer, Core Platform & Course Templating Concept',
  },
  JGStyle: {
    name: 'Josef Schmid',
    contribution: 'Student Data, Privacy & GDPR Compliance',
  },
  magkue: {
    name: 'Magnus Kühne',
    contribution: 'Developer Tooling & Testing Infrastructure',
  },
  robertjndw: {
    name: 'Robert Jandow',
    contribution: 'Server-side contributor',
  },
  maximiliansoelch: {
    name: 'Maximilian Sölch',
    contribution: 'Contributor',
  },
  bgeisb: {
    name: 'Benedikt Geisberger',
    contribution: 'UI/UX Design',
  },
  FelixTJDietrich: {
    name: 'Felix T.J. Dietrich',
    contribution: 'Intro Course Phase Maintainer',
  },
  phnagy: {
    name: 'Philipp Nagy',
    contribution: 'TEASE - Team Allocation Decision Support System',
  },
}
