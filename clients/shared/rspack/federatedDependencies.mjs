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
    SINGLETON_PACKAGES.map((name) => {
      const requiredVersion = packageJson.dependencies[name]
      if (!requiredVersion) {
        throw new Error(`federatedDependencies: "${name}" is not listed in clients/package.json`)
      }
      return [name, { singleton: true, requiredVersion }]
    }),
  )
