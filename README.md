### Setup Development Environment
1) Copy .env.example to .env
2) Fill out .env
3) Run command to create MariaDB Docker container
```shell
docker run --name=zephyr_mysql -e MYSQL_ROOT_PASSWORD=zephyr -e MYSQL_DATABASE=zephyr -p 3306:3306 -d mariadb
```

Install 'air' for auto reloads
```shell
go install github.com/cosmtrek/air@latest
```

## To-Do
- Consider if we should allow people to have multiple-levels to subdomains (ex: my.very.funny.domain.com)
  - Doing so would have implications regarding [SSL certs](https://developers.cloudflare.com/ssl/edge-certificates/advanced-certificate-manager)
- Implement [database locks](https://gorm.io/docs/advanced_query.html) to prevent concurrent modification issues
- Implement better error messages in handlers so user knows what's wrong (when possible).
- Popup for deleting Hosts in Dashboard that clarifies what it means (like GitHub does with archiving/deleting repos)
  - Images will NOT be deleted
  - Images will no longer be accessible from that hostname (perhaps paid customers can keep previous image links valid)
  - Future uploads to that hostname will fail
- Client IP address filtering for Cloudflare token for added security
- Cloudflare API is rate limited at 1200requests/5minutes/account. Ensure we can handle that.
  - Consider batching CNAME create/delete requests