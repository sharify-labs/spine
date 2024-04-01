### Setup Development Environment
1) Copy .env.example to .env
2) Fill out .env

Install 'air' for auto reloads
```shell
go install github.com/cosmtrek/air@latest
```

## To-Do
- Implement better error messages in handlers so user knows what's wrong (when possible).
- Popup for deleting Hosts in Dashboard that clarifies what it means (like GitHub does with archiving/deleting repos)
  - Images will NOT be deleted
  - Images will no longer be accessible from that hostname (perhaps paid customers can keep previous image links valid)
  - Future uploads to that hostname will fail
- Client IP address filtering for Cloudflare token for added security
- Cloudflare API is rate limited at 1200requests/5minutes/account. Ensure we can handle that.
  - Consider batching CNAME create/delete requests
- Consider using Hashicorp Vault for storing secrets