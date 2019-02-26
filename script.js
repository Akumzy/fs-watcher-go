const fs = require('fs').promises,
  { join } = require('path')
let dir = join(__dirname, './bin')
;(async () => {
  let files = await fs.readdir(dir)
  for (const file of files) {
    let newName
    if (file.includes('386')) {
      newName = file.replace('386', 'ia32')
    } else if (file.includes('amd64')) {
      newName = file.replace('amd64', 'x64')
    }
    await fs.rename(join(dir, file), join(dir, newName))
    console.log(`Renamed -> ${file} to ${newName}`)
  }
})()
