import CodeBlock from '../components/CodeBlock';

export default function AzureDocs() {
  return (
    <section id="backends-azure" className="py-10 border-b border-border">
      <h2 className="text-2xl font-bold text-text-heading mb-2">Azure Blob Storage</h2>
      <p className="text-text-muted mb-3">
        Azure backend lives in the <code className="text-accent font-mono">/azure</code> subpackage.
        Supports account-name/key, connection strings, and the default Azure credential chain
        (managed identity, environment variables, Azure CLI, etc.).
      </p>
      <code className="text-sm bg-bg-card px-2 py-1 rounded text-accent font-mono">
        import "github.com/KARTIKrocks/objstore/azure"
      </code>

      <h3 className="text-lg font-semibold text-text-heading mt-6 mb-2">Account name + key</h3>
      <CodeBlock code={`store, err := azure.New(ctx,
    azure.DefaultConfig().
        WithAccountName("myaccount").
        WithAccountKey("mykey").
        WithContainerName("mycontainer"),
)`} />

      <h3 className="text-lg font-semibold text-text-heading mt-8 mb-2">Connection string</h3>
      <CodeBlock code={`store, err := azure.New(ctx,
    azure.DefaultConfig().
        WithConnectionString("DefaultEndpointsProtocol=https;AccountName=...").
        WithContainerName("mycontainer"),
)`} />

      <h3 className="text-lg font-semibold text-text-heading mt-8 mb-2">Default Azure credentials</h3>
      <p className="text-text-muted mb-3">
        Falls back through managed identity → environment variables → Azure CLI:
      </p>
      <CodeBlock code={`store, err := azure.New(ctx,
    azure.DefaultConfig().
        WithAccountName("myaccount").
        WithContainerName("mycontainer"),
)`} />
    </section>
  );
}
