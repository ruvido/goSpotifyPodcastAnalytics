# ROADMAP

- [X] .env file
	x caddy log location
	x podcast rss url

- [ ] docker-compose.yml 
	- dockerfile with the compiled executable

- [ ] import caddy log data
	
# Options
- [X] option: --last #days (default 30)
- [ ] option: --filter (e.g. "s2") 

## Commands

- [ ] COMMAND streams
	- output: date (sort) | streams | # listeners 
	- timeline of streams and listeners

- [ ] COMMAND list
	- output: episode # | date (sort) | title | # streams | # streams (1st week) 
	- list all episodes in chronological order (can be filtered) 

- [ ] COMMAND summary 
	- output | streams             | all | spotify | webpage | other
	- output | number_of_listeners | all | spotify | webpage | other
	- summarized data for the show

## Notes
- streams   |  episode+ip+user_agent with size>0
- listeners |  ip+user_agent with size>0

## Improvements
- does it make sense to read everytime the entire json file? Maybe make a cached version of the digested data? Create a date based folder tree to look into the data?
