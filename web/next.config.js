/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: false,
  output: 'standalone',
  images: {
    unoptimized: true,
  },
  // trailingSlash: false,
  // cleanUrls: false,
  // distDir: 'dist',
};

module.exports = nextConfig;
