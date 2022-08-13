/* eslint-disable */
const nextSafe = require('next-safe')

const isDev = process.env.NODE_ENV !== 'production'

module.exports = {
  reactStrictMode: true,
  async headers () {
		return [
			{
				source: '/:path*',
				headers: nextSafe({ isDev }),
			},
		]
	},
}
