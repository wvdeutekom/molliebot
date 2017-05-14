# Molliebot

Try asking the molliebot what's for lunch today.


## Setup
Required environment variables:

* `API_KEY` - `default: empty`

Optional environment variables

* `CONFIG_LOCATION` - `default: './config.json'`
The complete filepath where the bot should look for a config file.
* `RESTRICT_TO_CONFIG_CHANNELS` - `default: false`
This sets wheter the bot should respond to any channel it is invited in (`true`) or respond only to channels it has been invited in _and_ are set in the config file in the `channels` array.
