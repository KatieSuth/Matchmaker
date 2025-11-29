Matchmaker is a free, open-source web application for organizing custom competitive games over Discord. It does this by allowing a game organizer to configure events that players can sign up to join, and it will ask players to provide their competitive rank (to be seen by the organizer but by default hidden to other players for privacy--can be configured in preferences). Once the organizer is ready, they can click a button to create 2 teams of the configured player count and sort players into those teams or a substitute pool as fairly as possibly based on their provided competitive ranks. It will also allow the organizer to make manual edits to the teams as needed.

It's still very much in the early phases but is intended to one day support Valorant and League of Legends, along with the ability for users to create their own custom settings for non-supported games or modify supported games (for example, 3v3 games instead of 5v5 games). See the roadmap for more information on where things are at & where we're going.

## Roadmap
- [x] Basic login functionality via Discord
- [x] User storage via SQLite
- [ ] Logout functionality
- [ ] User preferences form
- [ ] One-off event admin configuration
- [ ] One-off event sign-up for players
- [ ] One-off event admin team creation
- [ ] 1.0 web hosting with public availability
- [ ] Player duo requests
- [ ] Attendance tracking and host alerts for repetitive no-show players
- [ ] Riot API Linking for automatic competitive rank detection
- [ ] Tournament bracket event support
- [ ] Non-Riot game support (Overwatch 2, Marvel Rivals, etc.)


## FAQ
#### Why do I need this when Valorant custom games have an autobalance button?
The built-in team balancer is a great resource if you have exactly 10 players and just don't know how to configure the teams. If you have a pool of more than 10 players and could potentially be running multiple games at once though, it cannot be used to fairly determine who should be in which lobby. That said, this isn't just for enormous Discord servers that will have tens of players joining at once; it can also be used to determine good 2v2 or 3v3 matches from a pool of available competitors for, say, the new Skirmish mode in customs.

#### What if I want to double and triple check that the people who join my game are only from my Discord server?
The easiest way is to make sure you only share the sign-up code with players within your server. If you're concerned that someone in your server might be sharing the sign-up code with players outside of your server and you don't know how to stop them from doing that, you can self-host this application and configure it so that only players from your Discord server can log into it (details on how to self-host will be added after the 1.0 release); however, if this application ever gets users outside of just me, the developer, and this is something that is requested a lot, I will consider adding an event configuration setting to only allow players to join from specific Discord servers.

#### You've got a lot of roadmap there and not a lot of journey. Is this thing ever going to be done?
This application currently has one developer and I work on it when I can, so things might take a while. Thanks for the interest though and please feel free to add issues & contribute! I will review them as I'm able.


## Getting Started

First, run the development server:

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

This project uses [`next/font`](https://nextjs.org/docs/app/building-your-application/optimizing/fonts) to automatically optimize and load [Geist](https://vercel.com/font), a new font family for Vercel.
