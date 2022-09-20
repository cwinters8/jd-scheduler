# JD Scheduler

## Initial release goals

- [x] an admin should be able to create a new calendar with an arbitrary title
- [x] a user should be able to authenticate (via Stytch & Google) as a unique user
- [x] an admin should be able to add email addresses to a volunteer list. new members would be sent invites to login to the scheduler. only admin-provided emails should be able to login.
- [x] get auth working in the hosted app on replit
- [ ] fix role checks - should reference cockroach db instead of upstash redis
- [ ] configure oauth scopes for accessing the Calendar API
- [ ] handle user roles:
  - [x] admin
  - [x] volunteer
  - recruit?
    - may not actually need this?
      - do recruits need to login?
        - pros:
          - recruits could review and manage upcoming appointments
          - easy to convert to volunteers later
        - cons:
          - time ⏱
          - more stytch MAUs (which are cheap - 10c/user/month)
          - more users for admins to manage
- [ ] a volunteer should be able to add themself to a shift
- [ ] a recruit should be able to select a 15 min block of time from within scheduled shifts
  - [ ] the recruit should receive an email with a calendar invite

### Admin training

- [ ] how to set up shifts on the server owned Google Calendar

## Later goals

- [ ] an admin user should be able to schedule a shift via the user interface, rather than having to go to Google Calendar
- [ ] automated testing
