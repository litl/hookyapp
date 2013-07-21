# HookyApp - A simple webhook handler for HockeyApp

## Installation

Make sure you have your Go environment setup, in particular that GOPATH is set. Then run:

  $ go install github.com/litl/hookeyapp

## Configuration

You will need a hookeyapp.toml configuration file in the server's working directory. An
example is provided.

You will also need to configure your HockeyApp webhooks to point at
http://<host>:<port>/hockeyapp_webhook

## Copyright and License

HookyApp is Copyright (c) 2013 litl, LLC and licensed under the MIT license.
See the LICENSE file for full details.
