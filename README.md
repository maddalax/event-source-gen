This is a tool to generate event sourcing boilerplate code for a given domain.

**Usage:**

1. Edit the `config.yml` file at the root to define your events, commands, projections, and aggregates.
2. Run 'task run-gen' or 'task watch-gen' to generate the go code from your config.

Example Config:
```yaml
config:
  sourcing_dir: eventgen/internal/sourcing
  validations_dir: eventgen/internal/validators

aggregates:
    user:
      create:
        events: [ user.created ]
        fields:
          id: id [required]
          firstName: string
          lastName: string
          email: string

      update-email:
        events: [ user.email-changed ]
        fields:
          id: id [required]
          email: string

events:
    user:
      created:
        id: id
        firstName: string
        lastName: string
        email: string

      email-changed:
        id: id
        email: string

projections:
    emails-used:
      - user.created
      - user.email-changed

# Place any type mappings here.
# Example:
# type_mapping:
#   time:
#       type: time.Time
#       import: time
type_mapping:
    time:
        type: time.Time
        import: time
    core*:
        import: eventgen/internal/domain/core

# Place any events that you have renamed here.
# Example:
# event_migrations:
#   ContainerUserAssigned: [UserAssignedToContainer]
#
# This will map the event UserAssignedToContainer to ContainerUserAssigned event
event_migrations:
  UserCreated: [MyNewEvent]


projection_additional_methods:
  methods:
    SetLastEventVersion:
      params: 'version sourcing.Version'
    IsBehind:
      return: bool
    GetDiffBehind:
      return: sourcing.Version
    GetLastEventVersion:
      return: sourcing.Version
    Name:
      return: string
    IsHydrating:
      return: bool
    SetHydrating:
      params: 'hydrating bool'
    SetPaused:
      params: 'paused bool'
    IsPaused:
      return: bool
    IsCatchingUp:
      return: bool
    SetCatchingUp:
      params: 'catchingUp bool'
    Debug:
      return: bool
```
