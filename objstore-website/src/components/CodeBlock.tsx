import { useEffect, useState } from 'react';
import { getHighlighter } from '../lib/highlighter';
import { useTheme } from '../hooks/useTheme';

interface CodeBlockProps {
  code: string;
  lang?: string;
}

export default function CodeBlock({ code, lang = 'go' }: CodeBlockProps) {
  const [html, setHtml] = useState('');
  const [copied, setCopied] = useState(false);
  const { theme } = useTheme();

  const shikiTheme = theme === 'dark' ? 'github-dark' : 'github-light';

  useEffect(() => {
    getHighlighter().then((hl) => {
      setHtml(hl.codeToHtml(code.trim(), { lang, theme: shikiTheme }));
    });
  }, [code, lang, shikiTheme]);

  const handleCopy = () => {
    navigator.clipboard.writeText(code.trim());
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="relative group my-4">
      <button
        onClick={handleCopy}
        className="absolute top-2 right-2 p-1.5 rounded-md bg-overlay-hover text-text-muted hover:text-text opacity-0 group-hover:opacity-100 transition-opacity text-xs"
        aria-label="Copy code"
      >
        {copied ? 'Copied!' : 'Copy'}
      </button>
      {html ? (
        <div
          className="rounded-lg overflow-x-auto text-sm border border-border [&>pre]:p-4! [&>pre]:m-0! [&>pre]:rounded-lg! [&>pre]:bg-bg-card!"
          dangerouslySetInnerHTML={{ __html: html }}
        />
      ) : (
        <pre className="bg-bg-card rounded-lg p-4 text-sm overflow-x-auto">
          <code className="text-text-muted">{code.trim()}</code>
        </pre>
      )}
    </div>
  );
}
