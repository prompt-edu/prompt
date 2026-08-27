import path from 'node:path'
import { fileURLToPath } from 'node:url'
import rspack from '@rspack/core'
import { federatedDependencies } from './federatedDependencies.mjs'

const { ModuleFederationPlugin } = rspack.container

export const createRspackConfig = ({ name, port, configUrl, resolveAlias }) => {
  for (const [option, value] of Object.entries({ name, port, configUrl })) {
    if (!value) {
      throw new Error(`createRspackConfig: missing required option "${option}"`)
    }
  }

  const componentDir = path.dirname(fileURLToPath(configUrl))

  return (env = {}) => {
    const IS_DEV = env.NODE_ENV !== 'production'

    return {
      target: 'web',
      mode: IS_DEV ? 'development' : 'production',
      devtool: IS_DEV ? 'source-map' : undefined,
      entry: './src/index.js',
      devServer: {
        static: { directory: path.join(componentDir, 'public') },
        compress: true,
        hot: true,
        historyApiFallback: true,
        port,
        client: { progress: true },
        open: false,
      },
      module: {
        rules: [
          {
            test: /\.tsx?$/,
            use: {
              loader: 'builtin:swc-loader',
              options: {
                jsc: {
                  parser: { syntax: 'typescript', tsx: true },
                  transform: { react: { runtime: 'automatic' } },
                },
              },
            },
            exclude: /node_modules/,
          },
          {
            test: /\.css$/i,
            use: ['style-loader', 'css-loader', 'postcss-loader'],
            exclude: /node_modules/,
          },
          {
            test: /\.css$/i,
            include: /node_modules/,
            use: ['style-loader', 'css-loader'],
          },
        ],
      },
      output: {
        filename: '[name].[contenthash].js',
        path: path.resolve(componentDir, 'build'),
        publicPath: 'auto',
        clean: true,
      },
      resolve: {
        extensions: ['.ts', '.tsx', '.js', '.mjs', '.jsx'],
        ...(resolveAlias ? { alias: resolveAlias(componentDir) } : {}),
      },
      plugins: [
        new ModuleFederationPlugin({
          name,
          filename: 'remoteEntry.js',
          exposes: {
            './routes': './routes',
            './sidebar': './sidebar',
            './provide': './src/provide',
          },
          shared: federatedDependencies(),
        }),
        new rspack.CopyRspackPlugin({ patterns: [{ from: 'public' }] }),
        new rspack.HtmlRspackPlugin({
          template: 'public/template.html',
          minify: !IS_DEV,
        }),
      ],
      cache: { type: 'persistent' },
    }
  }
}
