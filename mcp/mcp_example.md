# 115driver MCP Server Example

This document provides examples of how to use the 115driver MCP server.

## Starting the Server

To start the MCP server, run:

```bash
go build -o mcp-server mcp/main.go
```

Then run the server with your 115 cookies:

```bash
./mcp-server --cookie="UID=your_uid;CID=your_cid;SEID=your_seid"
```

The server will listen on stdin/stdout for MCP requests.

Local file access is disabled by default. To allow `download_file` to write
locally, and to allow `upload_from_local` to read local files when destructive
tools are enabled, start the server with a local root; those tools can only read
or write paths under that directory:

```bash
./mcp-server --cookie="UID=your_uid;CID=your_cid;SEID=your_seid" --local-root="/safe/path"
```

MCP HTTP transfers default to a 2 hour total timeout. Override it with
`--download-timeout`, or use `--download-timeout=0` to disable the total timeout:

```bash
./mcp-server --cookie="UID=your_uid;CID=your_cid;SEID=your_seid" --download-timeout=6h
```

`download_file` has no size limit by default. `upload_from_url` defaults to a
2 GiB limit because it buffers through a local temp file. Use
`--download-max-bytes` or `--url-upload-max-bytes` to set limits; `0` disables
the corresponding size limit.
For SSRF protection, `upload_from_url` rejects non-HTTP(S) URLs, redirects to
unsafe hosts, and hostnames that resolve to loopback/private/link-local IPs.
When a hostname resolves to multiple safe IPs, MCP HTTP transfers try the later
addresses if earlier addresses are unreachable.

Tools that mutate 115 cloud state are not registered by default. Start the
server with `--allow-destructive-tools` to enable directory creation, file
rename/move/copy/delete, cloud uploads, recycle-bin mutations, and offline task
mutations:

```bash
./mcp-server --cookie="UID=your_uid;CID=your_cid;SEID=your_seid" --allow-destructive-tools
```

## Available Tools

### Account Tools

1. `getAccountInfo`: Get current account, storage space, and login device info
   - Parameters: none
   - Returns: `user`, `space.total`, `space.remain`, `space.used`, `login_devices`, and `imei_info`

### Directory Tools

1. `listDirectory`: List files and directories in a specific directory
   - Parameters:
     - `dir_id` (string): Directory ID to list, default is root directory (0)
     - `offset` (int64): Offset for pagination, default is 0
     - `limit` (int64): Number of items to return, default is 100, maximum is 500

2. `mkdir`: Create a new directory. Requires `--allow-destructive-tools`.
   - Parameters:
     - `parent_id` (string): Parent directory ID
     - `name` (string): Name of the new directory

### File Tools

1. `delete`: Delete files or directories. Requires `--allow-destructive-tools`.
   - Parameters:
     - `file_ids` (array of strings): IDs of files or directories to delete

2. `rename`: Rename a file or directory. Requires `--allow-destructive-tools`.
   - Parameters:
     - `file_id` (string): ID of file or directory to rename
     - `new_name` (string): New name for the file or directory

3. `move`: Move files or directories to another directory. Requires `--allow-destructive-tools`.
   - Parameters:
     - `dir_id` (string): Target directory ID
     - `file_ids` (array of strings): IDs of files or directories to move

4. `copy`: Copy files or directories to another directory. Requires `--allow-destructive-tools`.
   - Parameters:
     - `dir_id` (string): Target directory ID
     - `file_ids` (array of strings): IDs of files or directories to copy

5. `stat`: Get detailed information about a file or directory
   - Parameters:
     - `file_id` (string): ID of file or directory to get info

6. `download_file`: Download a file from 115 cloud storage to a local path.
   Requires `--local-root`.
   - Parameters:
     - `pick_code` (string): Pick code of the file to download
     - `local_path` (string): Local path under `--local-root`
     - `user_agent` (string): Optional User-Agent

7. `get_download_info`: Get a file download URL and metadata
   - Parameters:
     - `pick_code` (string): Pick code of the file
     - `user_agent` (string): Optional User-Agent

8. `upload_from_url`: Download a URL and upload it to 115 cloud storage.
   Requires `--allow-destructive-tools`.
   - Parameters:
     - `url` (string): HTTP or HTTPS URL to download
     - `dir_id` (string): Target 115 directory ID
     - `file_name` (string): Optional destination file name

9. `upload_from_local`: Upload a local file to 115 cloud storage.
   Requires `--allow-destructive-tools` and `--local-root`.
   - Parameters:
     - `local_path` (string): Local file path under `--local-root`
     - `dir_id` (string): Target 115 directory ID
     - `file_name` (string): Optional destination file name

### Recycle Bin Tools

1. `listRecycleBin`: List items in the recycle bin
   - Parameters:
     - `offset` (int): Offset for pagination, default is 0
     - `limit` (int): Number of items to return, default is 40, maximum is 100

2. `revertRecycleBin`: Revert items from the recycle bin. Requires `--allow-destructive-tools`.
   - Parameters:
     - `item_ids` (array of strings): IDs of items to revert

3. `cleanRecycleBin`: Clean items from the recycle bin. Requires `--allow-destructive-tools`.
   - Parameters:
     - `password` (string): Password for cleaning recycle bin
     - `item_ids` (array of strings): IDs of items to clean

### Share Tools

1. `getShareSnap`: Get shared files and directories snapshot information
   - Parameters:
     - `share_code` (string): Share code
     - `receive_code` (string): Receive code
     - `dir_id` (string): Directory ID to list, default is root directory

### Search Tools

1. `search`: Search for files and directories in the 115 cloud storage
   - Parameters:
     - `search_value` (string): Search keyword
     - `offset` (int): Offset for pagination, default is 0
     - `limit` (int): Limit number of results, default is 30
     - `type` (int): File type filter, 0:all 1:folder 2:document 3:image 4:video 5:audio 6:archive
     - `order` (string): Sort field, e.g. file_name, user_ptime
     - `asc` (int): Ascending order, 0:descending 1:ascending

### Offline Download Tools

1. `listOfflineTasks`: List offline download tasks
   - Parameters:
     - `page` (int64): Page number for pagination, default is 1

2. `addOfflineTaskURIs`: Add offline tasks by download URIs, supports http, ed2k, magnet. Requires `--allow-destructive-tools`.
   - Parameters:
     - `uris` (array of strings): Download URIs, supports http, ed2k, magnet
     - `save_dir_id` (string): Directory ID to save downloaded files

3. `deleteOfflineTasks`: Delete offline tasks. Requires `--allow-destructive-tools`.
   - Parameters:
     - `hashes` (array of strings): Task hashes to delete
     - `delete_files` (bool): Whether to delete associated files, default is false

4. `clearOfflineTasks`: Clear offline tasks. Requires `--allow-destructive-tools`.
   - Parameters:
     - `clear_flag` (int64): Clear flag, 0: clear completed tasks, 1: clear all tasks

## Example Request/Response

### Basic Directory Listing Request

```json
{
  "jsonrpc": "2.0",
  "id": "1",
  "method": "tools/call",
  "params": {
    "name": "listDirectory",
    "arguments": {
      "dir_id": "0"
    }
  }
}
```

### Basic Directory Listing Response

```json
{
  "jsonrpc": "2.0",
  "id": "1",
  "result": {
    "content": [
      {
        "type": "text",
        "text": "[{\"Id\":\"12345\",\"Name\":\"Documents\",\"Size\":0,\"Type\":1,\"CreateTime\":1234567890},{\"Id\":\"67890\",\"Name\":\"image.jpg\",\"Size\":1024,\"Type\":0,\"CreateTime\":1234567891}]"
      }
    ]
  }
}
```

### Search Request

Search for documents containing the word "report":

```json
{
  "jsonrpc": "2.0",
  "id": "2",
  "method": "tools/call",
  "params": {
    "name": "search",
    "arguments": {
      "search_value": "report",
      "limit": 10,
      "type": 2
    }
  }
}
```

### Search Response

```json
{
  "jsonrpc": "2.0",
  "id": "2",
  "result": {
    "content": [
      {
        "type": "text",
        "text": "{\"count\":5,\"files\":[{\"file_id\":\"12345\",\"name\":\"report.pdf\",\"size\":1024,\"is_directory\":false},{\"file_id\":\"12346\",\"name\":\"annual-report.docx\",\"size\":2048,\"is_directory\":false}],\"offset\":0,\"page_size\":10}"
      }
    ]
  }
}
```

## Notes

1. Valid cookies must be provided for authentication
2. The file list in the response is returned as a text content in JSON string format
3. The Type field indicates the file type: typically 1 for directories and 0 for files
4. All tools follow the standard MCP tool calling conventions
