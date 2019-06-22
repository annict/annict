<p align="center"><a href="https://annict.com" target="_blank" rel="noopener"><img src="https://user-images.githubusercontent.com/56767/56467671-fdd6ea80-645c-11e9-9056-a5d3fd5739e6.png" width="130" /></a></p>

# Annict (アニクト)

The platform for anime addicts.

[![CircleCI](https://img.shields.io/circleci/project/github/annict/annict.svg)](https://circleci.com/gh/annict/annict)
[![Test Coverage](https://api.codeclimate.com/v1/badges/ba10b596888853bc3f83/test_coverage)](https://codeclimate.com/github/annict/annict/test_coverage)
[![Code Climate](https://codeclimate.com/github/annict/annict/badges/gpa.svg)](https://codeclimate.com/github/annict/annict)
[![Hound](https://camo.githubusercontent.com/23ee7a697b291798079e258bbc25434c4fac4f8b/68747470733a2f2f696d672e736869656c64732e696f2f62616467652f50726f7465637465645f62792d486f756e642d6138373364312e737667)](https://houndci.com)
[![Discord](https://camo.githubusercontent.com/b12a95e20b7ca35f918c0ab5103fe56b6f44c067/68747470733a2f2f696d672e736869656c64732e696f2f62616467652f636861742d6f6e253230646973636f72642d3732383964612e737667)](https://discord.gg/PVJRUKP)


## Requirements

To run Annict on a local machine, you need to have the following dependencies installed:

- [Ruby](https://www.ruby-lang.org) 2.6.3
- [Docker](https://www.docker.com)
- [Docker Compose](https://docs.docker.com/compose/)


## Running the app

```
$ sudo sh -c "echo '127.0.0.1  annict.test' >> /etc/hosts"
$ sudo sh -c "echo '127.0.0.1  annict-jp.test' >> /etc/hosts"
$ git clone git@github.com:annict/annict.git
$ cd annict
$ bundle install
$ touch .env.development.local
$ bundle exec rails db:setup
$ docker-compose up --build
$ bundle exec rails s
```

You should then be able to open [http://annict.test:3000](http://annict.test:3000) (or [http://annict-jp.test:3000](http://annict-jp.test:3000) for Japanese) in your browser.


## Running Sorbet

We use [Sorbet](https://sorbet.org). You can run its commands like following:

```
$ ./bin/srb
$ ./bin/srb-rbi
```


## Running the tests

```
$ bundle exec rspec
```


### License

Copyright 2014-2019 Annict

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
