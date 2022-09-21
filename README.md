# JD Scheduler

## Initial release goals

- [x] an admin should be able to create a new calendar with an arbitrary title
- [x] a user should be able to authenticate (via Stytch & Google) as a unique user
- [x] an admin should be able to add email addresses to a volunteer list. new members would be sent invites to login to the scheduler. only admin-provided emails should be able to login.
- [x] get auth working in the hosted app on replit
- [x] fix role checks - should reference cockroach db instead of upstash redis
- [ ] configure stytch oauth scopes for accessing the Calendar API
- [ ] create a new calendar when db is initialized, and store its id etc in the db
  - need to figure out how to specify what users can manage events on the created calendar
- [ ] a volunteer should be able to add themself to a shift
- [ ] admins need to be able to be able to invite recruits
  - for now, recruits will be on a separate page in the admin portal. later, could create a common `users` UI to manage both types
- [ ] a recruit should be able to login
- [ ] a recruit should be able to select a 15 min block of time from within scheduled shifts
  - [ ] the recruit should receive an email with a calendar invite

## primetime requirements

- [ ] implement live/prod auth
- [ ] move mail config to justice dems domain
- [ ] connect Calendar API stuff to justice dems google workspace
- [ ] configure oauth consent screen in GCP (if not using OB GCP)
- [ ] clean database

### Admin training

- [ ] how to set up shifts on the server owned Google Calendar

## Later goals

- [ ] an admin user should be able to schedule a shift via the user interface, rather than having to go to Google Calendar
- [ ] automated testing
- [ ] admins need to be able to remove or update recruits and volunteers
- [ ] admins should be able to add/remove other admins
  - [ ] "root" admin (initial admin user created when project is first configured) should be protected from deletion
