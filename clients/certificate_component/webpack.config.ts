import path from 'path'
import 'webpack-dev-server'
import HtmlWebpackPlugin from 'html-webpack-plugin'
import packageJson from '../package.json' with { type: 'json' }
import webpack from 'webpack'
import container from 'webpack'
import { fileURLToPath } from 'url'
import CopyPlugin from 'copy-webpack-plugin'

const { ModuleFederationPlugin } = webpack.container

// ########################################
// ### Component specific configuration ###
// ########################################
const COMPONENT_NAME = 'certificate_component'
const COMPONENT_DEV_PORT = 3010

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

const config: (env: Record<string, string>) => container.Configuration = (env) => {
  const getVariable = (name: string) => env[name]

  const IS_DEV = getVariable('NODE_ENV') !== 'production'
  const deps = packageJson.dependencies

  return {
    target: 'web',
    mode: IS_DEV ? 'development' : 'production',
    devtool: IS_DEV ? 'source-map' : undefined,
    entry: './src/index.js',
    devServer: {
      static: {
        directory: path.join(__dirname, 'public'),
      },
      compress: true,
      hot: true,
      historyApiFallback: true,
      port: COMPONENT_DEV_PORT,
      client: {
        progress: true,
      },
      open: false,
    },
    module: {
      rules: [
        {
          test: /\.tsx?$/,
          use: 'ts-loader',
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
    },
    resolve: {
      extensions: ['.ts', '.tsx', '.js', '.mjs', '.jsx'],
      alias: {
        '@': path.resolve('../shared_library'),
      },
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
      new CopyPlugin({
        patterns: [{ from: 'public' }],
      }),
      new HtmlWebpackPlugin({
        template: 'public/template.html',
        minify: {
          removeComments: true,
          collapseWhitespace: true,
          removeRedundantAttributes: true,
          useShortDoctype: true,
          removeEmptyAttributes: true,
          removeStyleLinkTypeAttributes: true,
          keepClosingSlash: true,
          minifyJS: true,
          minifyCSS: true,
          minifyURLs: true,
        },
      }),
    ].filter(Boolean),
    cache: {
      type: 'filesystem',
    },
  }
}

export default config
