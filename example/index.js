const { default: Watcher, Op } = require('@akumzy/fs-watcher');
const { writeFileSync } = require('fs');
const w = new Watcher({
  binPath: '/home/akumzy/hola/fs-watcher-go/node-watcher',
  path: '/home/akumzy/hola/fs-watcher-go', // path you'll like to watch
  debug: true,
  interval: 1000,
  filters: [Op.Create, Op.Write, Op.Rename, Op.Move, Op.Remove], // changes to watch default is all
  recursive: true, // if the specified will be watch recursively or just is direct children
});
// start watching
w.start((err, files) => {
  if (err) {
    console.log(err);
    return;
  }
  let f = JSON.stringify(files, null, '   ');
  console.log(files.length);
  writeFileSync('./files.json', f);
});
w.onChange('create', file => {
  console.log(file);
});
w.onChange('write', file => {
  console.log(file);
});
w.onChange('rename', file => {
  console.log(file);
});
w.onAll((event, file) => {
  console.log(event, file);
});
w.onError(console.log);
