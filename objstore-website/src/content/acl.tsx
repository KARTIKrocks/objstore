interface AclRow {
  acl: string;
  description: string;
}

const rows: AclRow[] = [
  { acl: 'private', description: 'Owner-only access (default)' },
  { acl: 'public-read', description: 'Public read access' },
  { acl: 'public-read-write', description: 'Public read/write access' },
  { acl: 'authenticated-read', description: 'Authenticated users can read' },
  { acl: 'bucket-owner-full-control', description: 'Bucket owner has full control' },
];

export default function AclDocs() {
  return (
    <section id="acl" className="py-10 border-b border-border last:border-b-0">
      <h2 className="text-2xl font-bold text-text-heading mb-2">ACL Values</h2>
      <p className="text-text-muted mb-6">
        Common ACL values supported by S3 and GCS, used with <code className="font-mono">objstore.WithACL("...")</code>:
      </p>

      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-left">
              <th className="py-2 pr-4 text-text-heading font-semibold">ACL</th>
              <th className="py-2 text-text-heading font-semibold">Description</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={row.acl} className="border-b border-border/50">
                <td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">{row.acl}</td>
                <td className="py-2 text-text-muted">{row.description}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  );
}
