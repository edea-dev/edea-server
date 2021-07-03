# EDeA Front-End

Running the mock-up frontend standalone is not supported anymore. Run edead and you will see the frontend. 

To generate the necessary static files for the webserver:

```sh
./install-dependencies.sh
./build-fe.sh
```
## Requirements

Build process is only tested on linux, but might work with other versions of BASH, eg. Cygwin. You need to have a working install of 

 * yarn
 * BASH
 * sed

### Build options

You can modify default variables in the options.txt file. Its format is `VARIABLE="value"`, separated by newlines.

### Installing dependencies

To automatically install dependencies, just run `./install-dependencies.sh` - this will call yarn to pull down 3rd party modules like Bootstrap. This step should download all necessary resources from the internet.

### Creating public files

In order to make the frontend usable, a build script needs to copy the dependencies to the folder which is served by the http server. Just run `./build-fe.sh` to prepare the `public_html` folder.

### Worth mentioning

The file `resources.txt` lists every necessary file, line by line. Its format is `<source> <destination>\n`. You can use BASH globs (like `*` to match everything or `?` to match a single character). IMPORTANT: Do not use space in filenames!  
You must add any additional resource in this file. The sequence of files matter; you can overwrite resources with new ones if need be.

Yes, this process is cumbersome, since you need to add the same resource twice; yarn install (so the package.json file) and resources.txt - I wasn't able to find a more convenient way yet; there are complicated build systems out there with their own drawbacks.

After you change source files, you need to run `./build-fe.sh` again. The webserver does not need to be restarted. This step should not download new resources from the network and only use local files.

### Updating dependencies

Just run `yarn upgrade --latest`

## Post-processing

After copying file, a post-processing step is called; this is used to compress images, minify CSS, javascript, etc. It is stored in `post-process.sh` and is called by `build-fe.sh` automatically.

