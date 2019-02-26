const { default: Watcher, Op } = require('@akumzy/fs-watcher')

const w = new Watcher({
  path: '/home/akumzy/Documents', // path you'll like to watch
  debug: true,
  filters: [Op.Create, Op.Write, Op.Rename], // changes to watch default is all
  recursive: true // if the specified will be watch recursively or just is direct children
})
// start watching
w.start((err, files) => {
  if (err) {
    console.log(err)
    return
  }
  console.log('f', files)
  ;(async () => {
    try {
      let res = await w.ignore('/home/akumzy/Music/Culture')
      console.log('Ignore was ', res ? 'successful' : 'total bs')
      await w.addRecursive('/home/akumzy/Music')
    } catch (error) {
      console.log(error)
    }
  })()
})
w.onChange('create', file => {
  console.log(file)
})
w.onChange('write', file => {
  console.log(file)
})
w.onChange('rename', file => {
  console.log(file)
})
w.onAll((event, file) => {
  console.log(event, file)
})
w.onError(console.log)
setTimeout(async () => {
  let f = await w.getWatchedFiles()
  console.log(f)
}, 5000)
