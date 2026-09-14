import type { SidebarsConfig } from '@docusaurus/plugin-content-docs';

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

/**
 * Mirrors the flow of the repository README: what objstore is, how to
 * install it, one page per storage backend, then the operations and helpers
 * that work identically across all of them.
 */
const sidebars: SidebarsConfig = {
  docsSidebar: [
    'intro',
    'getting-started',
    {
      type: 'category',
      label: 'Backends',
      collapsed: false,
      items: [
        'backends/local',
        'backends/s3',
        'backends/gcs',
        'backends/azure',
        'backends/memory',
      ],
    },
    {
      type: 'category',
      label: 'Guides',
      collapsed: false,
      items: [
        'core-operations',
        'signed-urls',
        'helpers',
        'switching-backends',
        'errors',
      ],
    },
  ],
};

export default sidebars;
