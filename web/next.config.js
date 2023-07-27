/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: false,
  trailingSlash: true,
  output: 'export',
  images: {
    unoptimized: true,
  },
};

module.exports = nextConfig;
