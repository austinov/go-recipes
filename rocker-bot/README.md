# rocker-bot

Rocker is a slack bot. It shows calendar of rock-band's concerts and rock events in the cities.

![rocker bot](https://github.com/austinov/go-recipes/blob/assets/rocker-bot/screenshot.gif)


In the beginning, as usual, run:
```
    $ go get github.com/austinov/go-recipes
```

To run the rocker with the Docker use the script run-in-docker.sh.
Before running the script set token value of bot user in bot.yaml.
```
	$ cd github.com/austinov/go-recipes
	$ ./rocker-bot/run-in-docker.sh
```

The script creates two containers - first for PostgreSQL, second for the bot.
The loader uses the [www.concerts-metal.com](http://www.concerts-metal.com/) to load events.
With the current settings (in bot.yaml) the entire calendar is downloaded and available within the hour.
You can play with the settings of num-loaders and num-savers in bot.yaml.

To run the bot without using the Docker, create database structure with ./go-recipes/rocker-bot/sql/re-create-db
and specify the connection string to your PostgreSQL in bot.yaml and just run:
```
	$ cd github.com/austinov/go-recipes
	$ glide up
	$ cd ./rocker-bot
	$ go run main.go -config ./bot.yaml
```

To communicate with the bot you can use the following notation:

- to print help:
```
	@rocker help
```

- to list events of band:
```
	@rocker events of Metallica
```

- to list events in city:
```
	@rocker events in Paris
```

- to list events in city at the date (date format may be also dd.MM.yyyy or dd/MM/yyyy):
```
	@rocker events in London at 27 May 2017
	@rocker events in London at 15 Dec 2017
	@rocker events in London at 22.05.2017
	@rocker events in London at 17/05/2017
```

- to list events of band in city since the date:
```
	@rocker events of System of a Down in Dresden since 01 Jan 2017
```

- to list events in city till the date:
```
	@rocker events in Helsinki **till** 01 Jan 2017
```

- to list events in city since/till dates
```
	@rocker events in St Petersburg since 15 Dec 2016 till 01 Jan 2017
```

- to list events of band for period:
```
	@rocker events of Aerosmith for 15 Dec 2016 and 01 Jan 2017
```