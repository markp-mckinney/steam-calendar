# Steam Calendar

Query the Steam API to get wishlist items in ICS and JSON formats. The ICS file has all wishlist items that are still available. Items without release dates ("Coming Soon" or other) are set to 90 days from now. The JSON file is filtered to only upcoming releases sorted by release date followed by items without release dates.

Initially setup for usage with [homepage](https://github.com/gethomepage/homepage).

## Instructions

The script will write the ICS and JSON file for the given account and exit. To keep things up to date you can schedule it to run on a cron job.

### Environment Variables
|Variable|Example|Description|
|-|-|-|
|`OUT`|`/appdata/caddy/www/steamcal`|Folder which will be written to.|
|`STEAMID`|`12345678901234567`|Your Steam profile ID in steamID64 format (profile must be public).|

### Docker
Run with docker compose (or run equivalent). ex:
```
services:
  steam-calendar:
    container_name: steam-calendar
    image: ghcr.io/mpmckinney/steam-calendar:latest
    restart: no
    volumes:
      - /mnt/user/appdata/caddy/www/steamcal:/out
    environment:
      - STEAMID=12345678901234567
      - OUT=/out
networks: {}
```

### Using the output
The above examples are utilizing caddy to host the files which homepage then uses. Homepage usage:
```
- Media:
    - ' ':
        widget:
          type: calendar
          firstDayInWeek: sunday
          view: monthly
          maxEvents: 100
          showTime: true
          timezone: America/Los_Angeles
          integrations:
            - type: ical
              url: https://{{HOMEPAGE_VAR_DOMAIN}}/steamcal/wishlist.ics
              name: Steam Wishlist
              color: blue
    - 'Steam Upcoming':
        widget:
          type: customapi
          url: https://{{HOMEPAGE_VAR_DOMAIN}}/steamcal/upcoming.json
          display: dynamic-list
          mappings:
            items: items # optional, the path to the array in the API response. Omit this option if the array is at the root level
            name: name # required, field in each item to use as the item name (left side)
            label: releaseDate # required, field in each item to use as the item label (right side)
            format: date # optional - format of the label field
            limit: 100 # optional, limit the number of items to display
            target: https://store.steampowered.com/{storeUrlPath} # optional, makes items clickable with template support
```