# Annict

[![Travis CI](https://travis-ci.org/annict/annict.svg?branch=master)](https://travis-ci.org/annict/annict)
[![Code Climate](https://codeclimate.com/github/annict/annict/badges/gpa.svg)](https://codeclimate.com/github/annict/annict)
[![Coveralls](https://coveralls.io/repos/github/annict/annict/badge.svg?branch=master)](https://coveralls.io/github/annict/annict?branch=master)
[![Gemnasium](https://gemnasium.com/annict/annict.svg)](https://gemnasium.com/annict/annict)
[![Hound](https://camo.githubusercontent.com/23ee7a697b291798079e258bbc25434c4fac4f8b/68747470733a2f2f696d672e736869656c64732e696f2f62616467652f50726f7465637465645f62792d486f756e642d6138373364312e737667)](https://houndci.com)
[![Slack](https://slack.annict.com/badge.svg)](https://slack.annict.com)


**This branch is still under development for next version (V2).**

### Contributing

#### Requirements

To run Annict on local machine, you need to install some software below:

* Ruby 2.3
* PostgreSQL 9.5
* ImageMagick
* PhantomJS
  * For test

#### Running the app

```
$ git clone git@github.com:annict/annict.git
$ cd annict
$ cp config/application.yml{.example,}
$ bundle
$ rake db:setup
$ npm install
$ rails s -b 0.0.0.0
```

And you can access [http://localhost:3000](http://localhost:3000).

#### Running the test

```
$ rspec
```

### License

Copyright 2014-2016 Annict

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
