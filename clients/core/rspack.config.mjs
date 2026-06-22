import path from 'path'
import rspack from '@rspack/core'
import CompressionPlugin from 'compression-webpack-plugin'
import { BundleAnalyzerPlugin } from 'webpack-bundle-analyzer'
import packageJson from '../package.json' with { type: 'json' }
import { fileURLToPath } from 'url'

const { ModuleFederationPlugin } = rspack.container

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

const config = (env = {}) => {
  const IS_DEV = env.NODE_ENV !== 'production'
  const IS_PERF = env.BUNDLE_SIZE === 'true'
  const deps = packageJson.dependencies

  const templateURL = IS_DEV ? `http://localhost:3001` : `/template`
  const interviewURL = IS_DEV ? `http://localhost:3002` : `/interview`
  const matchingURL = IS_DEV ? `http://localhost:3003` : `/matching`
  const introCourseDeveloperURL = IS_DEV ? `http://localhost:3005` : `/intro-course-developer`
  const githubChallengeURL = IS_DEV ? `http://localhost:3006` : `/github-challenge`
  const assessmentURL = IS_DEV ? `http://localhost:3007` : `/assessment`
  const teamAllocationURL = IS_DEV ? `http://localhost:3008` : `/team-allocation`
  const selfTeamAllocationURL = IS_DEV ? `http://localhost:3009` : `/self-team-allocation`
  const certificateURL = IS_DEV ? `http://localhost:3010` : `/certificate`
  const infrastructureSetupURL = IS_DEV ? `http://localhost:3011` : `/infrastructure-setup`

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
        remotes: {
          template_component: `template_component@${templateURL}/remoteEntry.js?${Date.now()}`,
          interview_component: `interview_component@${interviewURL}/remoteEntry.js?${Date.now()}`,
          matching_component: `matching_component@${matchingURL}/remoteEntry.js?${Date.now()}`,
          intro_course_developer_component: `intro_course_developer_component@${introCourseDeveloperURL}/remoteEntry.js?${Date.now()}`,
          assessment_component: `assessment_component@${assessmentURL}/remoteEntry.js?${Date.now()}`,
          github_challenge_component: `github_challenge_component@${githubChallengeURL}/remoteEntry.js?${Date.now()}`,
          team_allocation_component: `team_allocation_component@${teamAllocationURL}/remoteEntry.js?${Date.now()}`,
          self_team_allocation_component: `self_team_allocation_component@${selfTeamAllocationURL}/remoteEntry.js?${Date.now()}`,
          certificate_component: `certificate_component@${certificateURL}/remoteEntry.js?${Date.now()}`,
          infrastructure_setup_component: `infrastructure_setup_component@${infrastructureSetupURL}/remoteEntry.js?${Date.now()}`,
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
    cache: { type: 'filesystem' },
  }
}

export default config
