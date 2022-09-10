# JD Scheduler

## Initial release goals

- [x] an admin should be able to create a new calendar with an arbitrary title
- [x] a user should be able to authenticate (via Stytch & Google) as a unique user
- [ ] figure out a way to approve users to access the server's calendar
      Options:
  - a staff member adding email addresses to the volunteer list. new members would be sent invites to login to the scheduler
  - ~~volunteer could create their account and then wait for approval~~ not a good option - delays would happen because staff would have to approve each individual account
- add ability for admin user to add volunteers to list and invite them
  - [ ] /admin route to list volunteers and invite new ones
- [ ] get auth working in the hosted app on repl.it
- [ ] configure oauth scopes for accessing the Calendar API
- [ ] handle user roles:
  - ~~admin~~ admin users will be handled later - for now everything they need to do will be via Google services manually
  - volunteer
  - recruit
- [ ] a volunteer should be able to add themself to a shift
- [ ] a recruit should be able to select a 15 min block of time from within scheduled shifts
  - [ ] the recruit should receive an email with a calendar invite

### Admin training

- [ ] how to set up shifts on the server owned Google Calendar

## Later goals

- [ ] an admin user should be able to schedule a shift via the user interface, rather than having to go to Google Calendar
