# Annict

[![CircleCI](https://img.shields.io/circleci/project/github/annict/annict.svg)](https://circleci.com/gh/annict/annict)
[![Test Coverage](https://api.codeclimate.com/v1/badges/ba10b596888853bc3f83/test_coverage)](https://codeclimate.com/github/annict/annict/test_coverage)
[![Code Climate](https://codeclimate.com/github/annict/annict/badges/gpa.svg)](https://codeclimate.com/github/annict/annict)
[![Hound](https://camo.githubusercontent.com/23ee7a697b291798079e258bbc25434c4fac4f8b/68747470733a2f2f696d672e736869656c64732e696f2f62616467652f50726f7465637465645f62792d486f756e642d6138373364312e737667)](https://houndci.com)
[![Discord](https://camo.githubusercontent.com/b12a95e20b7ca35f918c0ab5103fe56b6f44c067/68747470733a2f2f696d672e736869656c64732e696f2f62616467652f636861742d6f6e253230646973636f72642d3732383964612e737667)](https://chat.annict.com)


### Contributing

#### Requirements

To run Annict on a local machine, you need to have the following dependencies installed:

* Ruby 2.5
* PostgreSQL 10.4
* ImageMagick
* ChromeDriver

#### Running the app

```
$ git clone git@github.com:annict/annict.git
$ cd annict
$ cp config/application.yml{.example,}
$ bundle
$ rake db:setup
$ yarn
```

Run Rails server:

```
$ rails s -b 0.0.0.0
```

Start Forman to create asset files:

```
$ foreman start -f Procfile.dev
```

If you don't want to use Foreman, you need to run following commands:

```
$ yarn bsb:watch
$ yarn webpack:watch
```

You should then be able to open [http://localhost:3000](http://localhost:3000) in your browser.

#### Running the tests

```
$ rspec
```

### License

Copyright 2014-2018 Annict

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
