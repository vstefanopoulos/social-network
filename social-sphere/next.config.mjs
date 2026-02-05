/** @type {import('next').NextConfig} */
const nextConfig = {
  /* config options here */
  reactStrictMode: false,
  reactCompiler: true,
  output: 'standalone',
  images: {
    remotePatterns: [
      {
        protocol: 'http',
        hostname: 'localhost',
        port: '9000',
        pathname: '/uploads/**',
      },
      {
        protocol: 'http',
        hostname: 'localhost',
        port: '9000',
        pathname: '/uploads-variants/**',
      },
      // For internal Docker usage (when frontend runs in Docker)
      {
        protocol: 'http',
        hostname: 'minio',
        port: '9000',
        pathname: '/uploads/**',
      },
      {
        protocol: 'http',
        hostname: 'minio',
        port: '9000',
        pathname: '/uploads-variants/**',
      },
    ],
  },
};

export default nextConfig;
