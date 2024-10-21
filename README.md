# Terraform Provider FastLocal

This provider provides alternatives to the "local" provider that are much faster when dealing with many files.

The data sources in this provider affect your local filesystem. No, this is not how data sources should be used, but it is better in our specific use case.