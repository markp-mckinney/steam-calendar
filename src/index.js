import * as fs from 'fs'
import * as ics from 'ics'

const out = process.env.OUT
const steamid = process.env.STEAMID

if (out === undefined) {
  throw new Error('Missing OUT dir')
}

if (steamid === undefined) {
  throw new Error('Missing STEAMID')
}

function batch (arr, batchSize) {
  return arr.reduce((acc, _, index) => {
    if (index % batchSize === 0) {
      acc.push(arr.slice(index, index + batchSize))
    }
    return acc
  }, [])
}

async function getWishlistIds () {
  const wishlistResponse = await (await fetch(`https://api.steampowered.com/IWishlistService/GetWishlist/v1/?steamid=${steamid}`)).json()
  console.log(`Got ${wishlistResponse.response.items.length} wishlist items`)
  return wishlistResponse.response.items
}

async function getWishlistItems () {
  const wishlistIds = await getWishlistIds()

  const wishlistIdBatches = batch(wishlistIds, 200)

  const unfilteredItems = []
  const wishlistItems = []
  for (const wishlistIdBatch of wishlistIdBatches) {
    const input = {
      ids: wishlistIdBatch.map(item => ({ appid: item.appid })),
      context: {
        language: 'english',
        country_code: 'US',
        steam_realm: 1
      },
      data_request: {
        include_release: true,
        include_basic_info: true
      }
    }
    const batchUrl = `https://api.steampowered.com/IStoreBrowseService/GetItems/v1?input_json=${(JSON.stringify(input))}`
    const batchResult = await (await fetch(batchUrl)).json()
    const items = batchResult.response.store_items
    console.log(`Got data for ${items.length} items`)
    const filteredItems = items.filter(item => item.appid !== undefined && item.visible) // TODO: filter for future games
    console.log(`Filtered ${items.length - filteredItems.length} items`)
    wishlistItems.push(...filteredItems)
    unfilteredItems.push(...items)
  }

  write('unfiltered.json', JSON.stringify(unfilteredItems, null, 2), `${unfilteredItems.length} items`)
  write('filtered.json', JSON.stringify(wishlistItems, null, 2), `${wishlistItems.length} items`)

  return wishlistItems
}

async function createUpcomingJson (wishlistItems) {
  const knownDates = []
  const unknownDates = []
  wishlistItems.forEach(item => {
    if (item.release?.steam_release_date) {
      const releaseDate = new Date(item.release?.steam_release_date * 1000)
      if (releaseDate.getTime() > Date.now()) {
        knownDates.push({
          ...createItemJson(item, createDayArray(releaseDate).join('-')),
          rawDate: releaseDate.getTime()
        })
      }
    } else {
      unknownDates.push(createItemJson(item, item.release?.custom_release_date_message))
    }
  })
  const data = {
    items: [...(knownDates).sort((a, b) => a.rawDate - b.rawDate), ...unknownDates]
  }

  write('upcoming.json', JSON.stringify(data), `${data.items.length} items`)
}

function createItemJson (item, releaseDate) {
  return {
    name: item.name,
    releaseDate,
    storeUrlPath: item.store_url_path
  }
}

async function createIcs (wishlistItems) {
  const { value, error } = ics.createEvents(wishlistItems.map(item => {
    const { releaseDate, title, location } = getReleaseDateAndTitle(item)

    return {
      ...getFullDayStartAndEnd(releaseDate),
      title,
      description: item.basic_info?.short_description ?? '',
      uid: item.appid.toString(),
      url: `https://store.steampowered.com/${item.store_url_path}`,
      location // will be used as "additional" in homepage for hover text
      // type: item.type, // enum? 4 -> DLC
    }
  }))

  if (error) {
    console.error('Unable to create ics')
    throw error
  }

  write('wishlist.ics', value, 'ics')
}

function write (relativePath, file, customMessage) {
  const fullPath = `${out}/${relativePath}`
  console.log(`Writing ${customMessage ? `${customMessage} ` : ''}to ${fullPath}`)
  fs.writeFileSync(fullPath, file)
}

function getReleaseDateAndTitle (item) {
  if (item.release?.steam_release_date) {
    return {
      releaseDate: new Date(item.release.steam_release_date * 1000),
      title: item.name
    }
  }
  return {
    releaseDate: new Date(Date.now() + (oneDayMillis * 90)),
    title: `${item.name} ("${item.release?.custom_release_date_message}")`,
    location: item.release?.custom_release_date_message
  }
}

function createDayArray (date) {
  return [date.getFullYear(), date.getMonth() + 1, date.getDate()]
}

const oneDayMillis = 86400000
function getFullDayStartAndEnd (date) {
  return {
    start: createDayArray(date),
    end: createDayArray(new Date(date.getTime() + oneDayMillis))
  }
}

const wishlistItems = await getWishlistItems()
await createIcs(wishlistItems)
await createUpcomingJson(wishlistItems)
