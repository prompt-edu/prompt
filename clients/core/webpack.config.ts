import path from 'path'
import CompressionPlugin from 'compression-webpack-plugin'
import CopyPlugin from 'copy-webpack-plugin'
import 'webpack-dev-server'
import HtmlWebpackPlugin from 'html-webpack-plugin'
import CssMinimizerPlugin from 'css-minimizer-webpack-plugin'
import { BundleAnalyzerPlugin } from 'webpack-bundle-analyzer'
import { CleanWebpackPlugin } from 'clean-webpack-plugin'
import ExternalTemplateRemotesPlugin from 'external-remotes-plugin'
import packageJson from '../package.json' with { type: 'json' }
import { fileURLToPath } from 'url'
import container from 'webpack'
import webpack from 'webpack'

const { ModuleFederationPlugin } = webpack.container

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

// TODO: specify the version for react in shared dependencies
const config: (env: Record<string, string>) => container.Configuration = (env) => {
  const getVariable = (name: string) => env[name] // These variables are all determined at build time

  const IS_DEV = getVariable('NODE_ENV') !== 'production'
  const IS_PERF = getVariable('BUNDLE_SIZE') === 'true'
  const deps = packageJson.dependencies

  // Adjust this to match your deployment URL
  const templateURL = IS_DEV ? `http://localhost:3001` : `/template`
  const interviewURL = IS_DEV ? `http://localhost:3002` : `/interview`
  const matchingURL = IS_DEV ? `http://localhost:3003` : `/matching`
  const introCourseTutorURL = IS_DEV ? `http://localhost:3004` : `/intro-course-tutor`
  const introCourseDeveloperURL = IS_DEV ? `http://localhost:3005` : `/intro-course-developer`
  const devopsChallengURL = IS_DEV ? `http://localhost:3006` : `/devops-challenge`
  const assessmentURL = IS_DEV ? `http://localhost:3007` : `/assessment`
  const teamAllocationURL = IS_DEV ? `http://localhost:3008` : `/team-allocation`
  const selfTeamAllocationURL = IS_DEV ? `http://localhost:3009` : `/self-team-allocation`
  const certificateURL = IS_DEV ? `http://localhost:3010` : `/certificate`

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
      port: 3000,
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
          test: /\.css$/,
          include: [
            path.resolve(__dirname, 'src'),
            path.resolve(__dirname, '../node_modules/@xyflow/react/dist/style.css'),
            path.resolve(__dirname, '../node_modules/@tumaet/prompt-ui-components/dist'),
            path.resolve(__dirname, '../shared_library/components/minimal-tiptap/styles/index.css'),
          ],
          use: [
            'style-loader', // Injects styles into DOM
            'css-loader', // Resolves CSS imports
            'postcss-loader', // Processes Tailwind and other PostCSS plugins
          ],
        },
      ],
    },
    output: {
      filename: '[name].[contenthash].js',
      path: path.resolve(__dirname, 'build'),
      publicPath: '/',
    },
    resolve: {
      extensions: ['.ts', '.tsx', '.js', '.mjs', '.jsx'],
      alias: {
        '@': path.resolve(__dirname, '../shared_library'),
        '@core': path.resolve(__dirname, 'src'),
        '@managementConsole': path.resolve(__dirname, 'src/managementConsole'),
      },
    },
    plugins: [
      new ModuleFederationPlugin({
        name: 'core',
        remotes: {
          // The date will be resolved at buildtime and will force a cache reload after redeployment
          template_component: `template_component@${templateURL}/remoteEntry.js?${Date.now()}`,
          interview_component: `interview_component@${interviewURL}/remoteEntry.js?${Date.now()}`,
          matching_component: `matching_component@${matchingURL}/remoteEntry.js?${Date.now()}`,
          intro_course_tutor_component: `intro_course_tutor_component@${introCourseTutorURL}/remoteEntry.js?${Date.now()}`,
          intro_course_developer_component: `intro_course_developer_component@${introCourseDeveloperURL}/remoteEntry.js?${Date.now()}`,
          assessment_component: `assessment_component@${assessmentURL}/remoteEntry.js?${Date.now()}`,
          devops_challenge_component: `devops_challenge_component@${devopsChallengURL}/remoteEntry.js?${Date.now()}`,
          team_allocation_component: `team_allocation_component@${teamAllocationURL}/remoteEntry.js?${Date.now()}`,
          self_team_allocation_component: `self_team_allocation_component@${selfTeamAllocationURL}/remoteEntry.js?${Date.now()}`,
          certificate_component: `certificate_component@${certificateURL}/remoteEntry.js?${Date.now()}`,
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
      new ExternalTemplateRemotesPlugin(),
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
      new CopyPlugin({
        patterns: [{ from: 'public' }], // Copies env.js to your output root
      }),
      IS_PERF && new BundleAnalyzerPlugin(),
      new CleanWebpackPlugin(),
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
      runtimeChunk: {
        name: 'runtime',
      },
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
      minimizer: [`...`, new CssMinimizerPlugin()],
    },
    cache: {
      type: 'filesystem',
    },
  }
}

export default config
