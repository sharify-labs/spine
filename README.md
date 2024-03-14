### Setup Development Environment
1) Copy .env.example to .env
2) Fill out .env
3) Run command to create MariaDB Docker container
```bash
docker run --name=zephyr_mysql -e MYSQL_ROOT_PASSWORD=zephyr -e MYSQL_DATABASE=zephyr -p 3306:3306 -d mariadb
```

## To-Do
- Implement [database locks](https://gorm.io/docs/advanced_query.html) to prevent concurrent modification issues
- Implement better error messages in handlers so user knows what's wrong (when possible).
- Popup for deleting Hosts in Dashboard that clarifies what it means (like GitHub does with archiving/deleting repos)
  - Images will NOT be deleted
  - Images will no longer be accessible from that hostname (perhaps paid customers can keep previous image links valid)
  - Future uploads to that hostname will fail
- Client IP address filtering for Cloudflare token for added security