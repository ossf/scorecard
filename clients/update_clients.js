const fs = require('fs');
const files = ['gitlabrepo/client.go', 'localdir/client.go', 'git/client.go', 'azuredevopsrepo/client.go', 'ossfuzz/client.go'];

files.forEach(f => {
  try {
    let c = fs.readFileSync(f, 'utf8');
    c = c.replace(/func \(client \*Client\) IsArchived\(\) \(bool, error\) \{[\s\S]*?\}/, match => {
      return match + '\n\n// IsPrivateVulnerabilityReportingEnabled implements RepoClient.IsPrivateVulnerabilityReportingEnabled.\nfunc (client *Client) IsPrivateVulnerabilityReportingEnabled() (bool, error) {\n\treturn false, fmt.Errorf("IsPrivateVulnerabilityReportingEnabled: %w", clients.ErrUnsupportedFeature)\n}';
    });
    fs.writeFileSync(f, c);
    console.log('Updated ' + f);
  } catch(e) {
    console.log('Error on ' + f + ': ' + e.message);
  }
});
