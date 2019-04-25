# Confluence Space PDF Export
Export Confluence spaces as PDF using this command-line tool.
This tool uses the SOAP API in Confluence as the REST API
does not support export to PDF.

## Requirements
- In order to enable the SOAP API in Confluence, enable the system plug-in/app `Confluence Axis SOAP plugin`. 
- Give the user `Space Export` permissions on the spaces you want to export with this tool.

## Hot to use?
When invoking the CLI you have to provide som information, the following
arguments are required:

- `-server` The FQDN for your Confluence Server, e.g. https://confluence.domain.com
- `-username` Username for the user performing the export.
- `-password` Password for the user performing the export.
- `-spaceKey` Confluence space key for the space to export.
- `-exportDirectory` File path to directory which PDF export is stored.
