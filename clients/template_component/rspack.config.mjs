import path from 'node:path'
import { fileURLToPath } from 'node:url'
import rspack from '@rspack/core'
import packageJson from '../package.json' with { type: 'json' }

const { ModuleFederationPlugin } = rspack.container

const COMPONENT_NAME = 'template_component'
const COMPONENT_DEV_PORT = 3001

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

const config = (env = {}) => {
  const IS_DEV = env.NODE_ENV !== 'production'
  const deps = packageJson.dependencies

  return {
    target: 'web',
    mode: IS_DEV ? 'development' : 'production',
    devtool: IS_DEV ? 'source-map' : undefined,
    entry: './src/index.js',
    devServer: {
      static: { directory: path.join(__dirname, 'public') },
      compress: true,
      hot: true,
      historyApiFallback: true,
      port: COMPONENT_DEV_PORT,
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
      path: path.resolve(__dirname, 'build'),
      publicPath: 'auto',
      clean: true,
    },
    resolve: {
      extensions: ['.ts', '.tsx', '.js', '.mjs', '.jsx'],
    },
    plugins: [
      new ModuleFederationPlugin({
        name: COMPONENT_NAME,
        filename: 'remoteEntry.js',
        exposes: {
          './routes': './routes',
          './sidebar': './sidebar',
          './provide': './src/provide',
        },
        shared: {
          react: { singleton: true, requiredVersion: deps.react },
          'react-dom': { singleton: true, requiredVersion: deps['react-dom'] },
          'react-router-dom': { singleton: true, requiredVersion: deps['react-router-dom'] },
          '@tanstack/react-query': {
            singleton: true,
            requiredVersion: deps['@tanstack/react-query'],
          },
          '@tumaet/prompt-shared-state': {
            singleton: true,
            requiredVersion: deps['@tumaet/prompt-shared-state'],
          },
        },
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

export default config
