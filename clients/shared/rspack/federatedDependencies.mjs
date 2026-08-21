import packageJson from '../../package.json' with { type: 'json' }

const SINGLETON_PACKAGES = [
  'react',
  'react-dom',
  'react-router-dom',
  '@tanstack/react-query',
  '@tumaet/prompt-shared-state',
]

export const federatedDependencies = () =>
  Object.fromEntries(
    SINGLETON_PACKAGES.map((name) => [
      name,
      { singleton: true, requiredVersion: packageJson.dependencies[name] },
    ]),
  )
