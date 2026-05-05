import { useState } from 'react';
import ThemeProvider from './components/ThemeProvider';
import Navbar from './components/Navbar';
import Sidebar from './components/Sidebar';
import Hero from './components/Hero';
import GettingStarted from './content/getting-started';
import LocalDocs from './content/local';
import S3Docs from './content/s3';
import GcsDocs from './content/gcs';
import AzureDocs from './content/azure';
import MemoryDocs from './content/memory';
import OperationsDocs from './content/operations';
import HelpersDocs from './content/helpers';
import SwitchingDocs from './content/switching';
import ErrorsDocs from './content/errors';
import AclDocs from './content/acl';

export default function App() {
  const [menuOpen, setMenuOpen] = useState(false);

  return (
    <ThemeProvider>
      <div className="min-h-screen">
        <Navbar onMenuToggle={() => setMenuOpen((o) => !o)} menuOpen={menuOpen} />
        <Sidebar open={menuOpen} onClose={() => setMenuOpen(false)} />

        <main className="pt-16 md:pl-64">
          <div className="max-w-4xl mx-auto px-4 md:px-8 pb-20">
            <Hero />
            <GettingStarted />
            <LocalDocs />
            <S3Docs />
            <GcsDocs />
            <AzureDocs />
            <MemoryDocs />
            <OperationsDocs />
            <HelpersDocs />
            <SwitchingDocs />
            <ErrorsDocs />
            <AclDocs />

            <footer className="py-10 text-center text-sm text-text-muted border-t border-border mt-10">
              <p>
                objstore is open source under the{' '}
                <a
                  href="https://github.com/KARTIKrocks/objstore/blob/main/LICENSE"
                  className="text-primary hover:underline"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  MIT License
                </a>
              </p>
            </footer>
          </div>
        </main>
      </div>
    </ThemeProvider>
  );
}
