import fs from 'node:fs';
import path from 'node:path';
import type * as Preset from '@docusaurus/preset-classic';
import type { Config } from '@docusaurus/types';
import { themes as prismThemes } from 'prism-react-renderer';

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

/**
 * Versioning policy — see VERSIONING.md for the full runbook.
 *
 * objstore is still pre-1.0, so versioning stays off until v1.0.0 (any 0.x
 * release can change documented behaviour under semver, which would make
 * snapshotting per-minor pointless — see VERSIONING.md). versions.json is
 * empty until the first `npm run cut-version` call, at which point `docs/`
 * automatically starts serving at /docs/next/ instead of /docs/ — no config
 * change needed here.
 */
const MAX_LIVE_VERSIONS: number = JSON.parse(
  fs.readFileSync(path.resolve(__dirname, 'versions.config.json'), 'utf8'),
).maxLiveVersions;

const versionsFile = path.resolve(__dirname, 'versions.json');
const allVersions: string[] = fs.existsSync(versionsFile)
  ? JSON.parse(fs.readFileSync(versionsFile, 'utf8'))
  : [];

/**
 * `DOCS_FAST_BUILD=true` builds only the in-progress docs. Used by `start` and
 * by the PR build check, where rebuilding every historical version is wasted
 * work once snapshots exist. Production deploys build the full live set.
 */
const fastBuild = process.env.DOCS_FAST_BUILD === 'true';
const liveVersions = allVersions.slice(0, MAX_LIVE_VERSIONS);
const includedVersions = fastBuild
  ? ['current', ...allVersions.slice(0, 1)]
  : ['current', ...liveVersions];

const config: Config = {
  title: 'objstore',
  tagline: 'Unified file storage interface for Go',
  // .ico carries 16/32/48 frames for the browsers that ignore SVG favicons;
  // the SVG below is preferred by everything current and stays sharp on hidpi.
  favicon: 'img/favicon.ico',

  headTags: [
    {
      tagName: 'link',
      attributes: {
        rel: 'icon',
        type: 'image/svg+xml',
        href: '/objstore/img/favicon.svg',
      },
    },
    {
      tagName: 'link',
      attributes: {
        rel: 'apple-touch-icon',
        href: '/objstore/img/apple-touch-icon.png',
      },
    },
  ],

  future: {
    v4: true,
    // Rspack/SWC build pipeline.
    faster: true,
  },

  url: 'https://kartikrocks.github.io',
  baseUrl: '/objstore/',

  organizationName: 'KARTIKrocks',
  projectName: 'objstore',
  trailingSlash: false,

  onBrokenLinks: 'throw',

  markdown: {
    hooks: {
      onBrokenMarkdownLinks: 'throw',
    },
  },

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          editUrl: 'https://github.com/KARTIKrocks/objstore/tree/main/website/',
          routeBasePath: 'docs',
          // Only overridden once a snapshot exists (see VERSIONING.md) — until
          // then `current` serves at /docs/ directly, with no /docs/next/ split.
          ...(allVersions.length > 0 && {
            versions: {
              current: {
                label: 'Next (unreleased)',
                path: 'next',
                banner: 'unreleased',
              },
            },
          }),
          onlyIncludeVersions: includedVersions,
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
        sitemap: {
          lastmod: 'date',
          changefreq: 'weekly',
          priority: 0.5,
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    image: 'img/objstore-social-card.png',
    colorMode: {
      defaultMode: 'light',
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'objstore',
      logo: {
        alt: 'objstore',
        src: 'img/logo.svg',
        // The mark uses the same primary token as the theme, so it needs the
        // dark-mode value (#6366f1) on a slate ground the way every other
        // primary-colored element does.
        srcDark: 'img/logo-dark.svg',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docsSidebar',
          position: 'left',
          label: 'Docs',
        },
        // Appears once the first snapshot is cut (see VERSIONING.md); an empty
        // versions.json means there is nothing yet to pick between.
        ...(allVersions.length > 0
          ? [
              {
                type: 'docsVersionDropdown' as const,
                position: 'right' as const,
                dropdownItemsAfter: [
                  {
                    href: 'https://github.com/KARTIKrocks/objstore/releases',
                    label: 'All releases',
                  },
                ],
              },
            ]
          : []),
        {
          href: 'https://pkg.go.dev/github.com/KARTIKrocks/objstore',
          label: 'API Reference',
          position: 'right',
        },
        {
          href: 'https://github.com/KARTIKrocks/objstore',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'light',
      links: [
        {
          title: 'Docs',
          items: [
            { label: 'Getting Started', to: '/docs/getting-started' },
            { label: 'Backends', to: '/docs/backends/local' },
            { label: 'Core Operations', to: '/docs/core-operations' },
          ],
        },
        {
          title: 'Reference',
          items: [
            {
              label: 'pkg.go.dev',
              href: 'https://pkg.go.dev/github.com/KARTIKrocks/objstore',
            },
            {
              label: 'Changelog',
              href: 'https://github.com/KARTIKrocks/objstore/blob/main/CHANGELOG.md',
            },
            {
              label: 'Releases',
              href: 'https://github.com/KARTIKrocks/objstore/releases',
            },
          ],
        },
        {
          title: 'More',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/KARTIKrocks/objstore',
            },
            {
              label: 'Issues',
              href: 'https://github.com/KARTIKrocks/objstore/issues',
            },
            {
              label: 'Contributing',
              href: 'https://github.com/KARTIKrocks/objstore/blob/main/CONTRIBUTING.md',
            },
          ],
        },
      ],
      copyright: `objstore is open source under the MIT License. Copyright © ${new Date().getFullYear()}.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['go', 'bash', 'json', 'yaml'],
    },
    // Algolia DocSearch — apply at https://docsearch.algolia.com/apply/
    // Uncomment and fill in the credentials once the application is approved.
    // algolia: {
    //   appId: 'YOUR_APP_ID',
    //   apiKey: 'YOUR_SEARCH_API_KEY',
    //   indexName: 'objstore',
    //   contextualSearch: true,
    // },
  } satisfies Preset.ThemeConfig,
};

export default config;
