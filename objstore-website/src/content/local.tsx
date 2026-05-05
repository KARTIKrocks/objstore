import CodeBlock from '../components/CodeBlock';

export default function LocalDocs() {
  return (
    <section id="backends" className="py-10 border-b border-border">
      <h2 className="text-2xl font-bold text-text-heading mb-2">Storage Backends</h2>
      <p className="text-text-muted mb-8">
        Every backend implements the same <code className="text-accent font-mono">objstore.Storage</code> interface,
        so application code stays the same regardless of where files actually live.
      </p>

      <div id="backends-local">
        <h3 className="text-xl font-semibold text-text-heading mb-2">Local Filesystem</h3>
        <p className="text-text-muted mb-4">
          Stores files on the local disk. Configurable base path, public URL prefix, and
          file/directory permissions.
        </p>

        <CodeBlock code={`config := objstore.LocalConfig{
    BasePath:        "./storage",
    BaseURL:         "https://example.com/files",
    CreateDirs:      true,
    FilePermissions: 0644,
    DirPermissions:  0755,
}

store, err := objstore.NewLocalStorage(config)`} />

        <p className="text-text-muted mt-4 mb-3">Or with the builder pattern:</p>
        <CodeBlock code={`store, err := objstore.NewLocalStorage(
    objstore.DefaultLocalConfig().
        WithBasePath("/var/uploads").
        WithBaseURL("https://cdn.example.com"),
)`} />
      </div>
    </section>
  );
}
