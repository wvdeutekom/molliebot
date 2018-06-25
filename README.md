# Molliebot

Try asking the molliebot what's for lunch or who is on-call today.


## Setup
Environment variables:

| Variable                      | Required | Default         | Description                                                                                                                                                                                   |
| :---                          | :---:    | :---            | :---                                                                                                                                                                                          |
| `API_KEY`                     | Yes      |                 | Slack API key                                                                                                                                                                                 |
| `PAGERDUTY_API_KEY`           | Yes      |                 | Pagerduty API key                                                                                                                                                                             |
| `DEBUG`                       | No       | 'false'         | Enables or disables full debug log of the slack API in stdout                                                                                                                                 |
| `CONFIG_LOCATION`             | No       | './config.json' | The complete filepath where the bot should look for a config file.                                                                                                                            |
| `RESTRICT_TO_CONFIG_CHANNELS` | No       | 'false'         | This sets wheter the bot should respond to any channel it is invited in (`true`) or respond only to channels it has been invited in _and_ are set in the config file in the `channels` array. |
|                               |          |                 |                                                                                                                                                                                               |
## Development

First, install all the dependencies

    glide update

Then run make dev through the Makefile

    make dev

## Building and deployment
Requirements:
* [Expenv](https://github.com/blang/expenv)

All scripts are located in the 'scripts/' directory. To build a new image you can run:

    ./build.sh

To build and immediately upload a new image to the docker hub and tag both `latest` and the current `gitref` use:

    ./build.sh upload

In order to deploy to the kubernetes cluster make sure you have the right cluster selected in your kube config and run:

    ./deploy.sh production

The parameter `production` can be replaced with `development`. For now the only difference in the state of the `DEBUG` env variable, for development it has been turned on.

Variables in the `kubernetes.yml` are supported through [Expenv](https://github.com/blang/expenv).
