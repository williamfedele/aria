_This is a place to keep track of what's going on and what I'm trying to do._

### 2024/8/15

Tried to create an initial menu with the choice to load the library list to play music. Right now the tracklist UI receives the tracks from main.go to populate its list. I can create with the menu and create the library list at the same time (since the initial NewModel function is the only receiver of the track list from main.go). But to return from the library to the menu, it expects the track list again. The only choice I saw was to load the library within the UI which I don't like. I'd also like to avoid using globals but this may be the easiest solution. Delayed this until I can figure out a cleaner way to handle this.

Tried to figure out adding additional keymaps to the list.Model but the delegate key stuff got complicated really fast. Also, using a custom delegate means I don't have immediate access to the trackListModel to send the chosen song into the trackFeed channel. I need to look more into how it all works so I can work up minimally.

### 2024/8/16

Transferred the UI to bubbletea list because it comes batteries included with pagination, filtering, and sensible default keybinds for list navigation. Spent a solid few hours trying to debug why the UI would completely glitch out when filtering. Turns it it was just Korean characters in some track titles that wouldn't render correctly so filtering would throw elements all over the terminal. Fun. Next to do is the ability to pause playback that was bugging out a few days ago. I would also like to let the player send updates back to the model to keep the UI and Player in sync for play/pause events.