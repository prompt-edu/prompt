import path from 'node:path'
import { fileURLToPath } from 'node:url'
import rspack from '@rspack/core'
import CompressionPlugin from 'compression-webpack-plugin'
import { BundleAnalyzerPlugin } from 'webpack-bundle-analyzer'
import { federatedDependencies } from '../shared/rspack/federatedDependencies.mjs'
import { resolveRemotes } from './remotes.config.mjs'

const { ModuleFederationPlugin } = rspack.container

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

const config = (env = {}) => {
  const IS_DEV = env.NODE_ENV !== 'production'
  const IS_PERF = env.BUNDLE_SIZE === 'true'

  const remotes = resolveRemotes(IS_DEV)

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
      port: 3000,
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
          test: /\.css$/,
          include: [
            path.resolve(__dirname, 'src'),
            path.resolve(__dirname, '../node_modules/@xyflow/react/dist/style.css'),
          ],
          use: ['style-loader', 'css-loader', 'postcss-loader'],
        },
        {
          test: /\.css$/,
          include: [path.resolve(__dirname, '../node_modules/@tumaet/prompt-ui-components/dist')],
          use: ['style-loader', 'css-loader'],
        },
      ],
    },
    output: {
      filename: '[name].[contenthash].js',
      path: path.resolve(__dirname, 'build'),
      publicPath: '/',
      clean: true,
    },
    resolve: {
      extensions: ['.ts', '.tsx', '.js', '.mjs', '.jsx'],
      alias: {
        '@core': path.resolve(__dirname, 'src'),
        '@managementConsole': path.resolve(__dirname, 'src/managementConsole'),
      },
    },
    plugins: [
      new ModuleFederationPlugin({
        name: 'core',
        remotes: Object.fromEntries(
          remotes.map(({ name, url }) => [name, `${name}@${url}/remoteEntry.js?${Date.now()}`]),
        ),
        shared: federatedDependencies(),
      }),
      new rspack.DefinePlugin({ __PROMPT_REMOTES__: JSON.stringify(remotes) }),
      new rspack.HtmlRspackPlugin({
        template: 'public/template.html',
        minify: !IS_DEV,
      }),
      new rspack.CopyRspackPlugin({
        patterns: [{ from: 'public' }],
      }),
      IS_PERF && new BundleAnalyzerPlugin(),
      !IS_DEV &&
        new CompressionPlugin({
          filename: '[path][base].gz',
          algorithm: 'gzip',
          test: /\.(js|css|html|svg)$/,
          threshold: 10240,
          minRatio: 0.8,
        }),
    ].filter(Boolean),
    optimization: {
      minimize: !IS_DEV,
      runtimeChunk: { name: 'runtime' },
      splitChunks: {
        chunks: 'async',
        minSize: 30000,
        minChunks: 1,
        maxAsyncRequests: 5,
        maxInitialRequests: 3,
        cacheGroups: {
          default: {
            name: 'common',
            chunks: 'initial',
            minChunks: 2,
            priority: -20,
            reuseExistingChunk: true,
          },
          vendors: {
            test: /[\\/]node_modules[\\/]/,
            name: 'vendors',
            chunks: 'all',
            priority: 10,
          },
        },
      },
      minimizer: ['...', new rspack.LightningCssMinimizerRspackPlugin()],
    },
    cache: { type: 'persistent' },
  }
}

export default config
