# EDeA Front-End

## TODO list / NOTES for INTERNAL USE
Delete this section before public release. TODO list.
 * license in the package.json is currently set to GPLv2
 * in resources.txt change the `*.*` globs to `*.min.css/js`
 * configure post-processing (css-js minification)
 * get rid of unnecessary parts of bootstrap-table

## Development

tl;dr `./quickstart.sh` 

### Requirements

Build process is only tested on linux, but might work with other versions of BASH, eg. Cygwin. You need to have a working install of 

 * npm
 * BASH
 * sed
 * wget
 * A webserver of your choice or Python (using Python3's http.server module)

### Build options

You can modify default variables in the options.txt file. Its format is `VARIABLE="value"`, separated by newlines.

### Installing dependencies

To automatically install dependencies, just run `./install-dependencies.sh` - this will call npm to pull down 3rd party modules like Bootstrap. This step should download all necessary resources from the internet.

### Creating public files

In order to make the frontend usable, a build script needs to copy the dependencies to the folder which is served by the http server. Just run `./build-fe.sh` to prepare the `public_html` folder.

### Serving files

Thanks to recent developments in web security, files served through the `file://` local protocol are no longer treated as equals, which requires a local webserver in order to reduce headache. You can use any webserver of your choice, or just run `./serve.sh` in order to start a simple Python http server. This should be built into Python. Now you can navigate to [http://localhost:8008/] to try the frontend. The Python http server keeps running until you hit Ctrl-C.

### Worth mentioning

The file `resources.txt` lists every necessary file, line by line. Its format is `<source> <destination>\n`. You can use BASH globs (like `*` to match everything or `?` to match a single character). IMPORTANT: Do not use space in filenames!  
You must add any additional resource in this file. The sequence of files matter; you can overwrite resources with new ones if need be.

Yes, this process is cumbersome, since you need to add the same resource twice; npm install (so the package.json file) and resources.txt - I wasn't able to find a more convenient way yet.

After you change source files, you need to run `./build-fe.sh` again. The webserver does not need to be restarted. This step should not download new resources from the network and only use local files.

#### Post-processing

After copying file, a post-processing step is called; this is used to compress images, minify CSS, javascript, etc. It is stored in `post-process.sh` and is called by `build-fe.sh` automatically.

## Q&A
### This looks rather convoluted and reminds me of a Rube Goldberg machine.
Welcome to web development.

### Why are you not using Grunt?
I tried but it is so badly documented that after spending 3 hours on figuring it out, I decided to just write what I need in BASH. Feel free to reimplement it with Grunt and submit a pull request. Also if you got the magic tool that *Just Does the Job*, please tell us.

### npm is evil!
I don't like it much either, but it is more convenient than hunting down resources by hand.

## Concepts
Nothing more than a lot of research, trial-and-error (mostly error) at this point, but hopefully it will gain some shape soon.
