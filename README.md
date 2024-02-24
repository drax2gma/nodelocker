# NodeLocker

## Building the project

As this project was developed and tested on x86_64 linux, I'm providing a simple shell script to make binary building easier. Please see `./build_linux.sh` in the project's root directory.

## Running tests

> ⚠️ Please be aware that running the included ShellSpec tests **WILL DESTROY ALL YOUR LOCAL REDIS DATABASE**!
> Use it with extra caution!

The test definitions are inside `tests/full_spec.sh`

You can read more on ShellSpec format here: https://shellspec.info/

Run the tests from the project directory with the following command:

```bash
❯ shellspec tests/
```

## Running NodeLocker

After compiling with `./build_linux.sh` the output binary will be at `bin/release/nodelocker-linux`. When the local Redis database is ready, just run the compiled binary from there manually for testing.

```bash
❯ ./nodelocker-linux
```

Of course, the _lab-prod_ behavior needs some extra setup, like a _systemd_ module. I prefer _supervisord_ for running such user-mode applications, your mileage may vary.

Just keep in mind, that if somehow the app fails, it won't restart itself, there is no watchdog feature implemented.

## Getting started

## Using NodeLocker

NodeLocker is a REST-like API service that is controllable via GET requests. For the sake of simplicity, POST, PUT and other methods are not used.

### Administrator functions

General request format:

```bash
❯ https://example.local:3000/admin?action=<some_action>&name=<entity_name>&token=<admin_token>
```

**<admin_token>** is the passphrase token for the admin user. To generate a token, you can simply create it with some hashing command, like this:

```bash
❯ echo "my beautiful password" | sha256sum | cut -d" " -f1
```

The bash command will return something like this:

```bash
5c8ccefbcb9f9a0ef8a8d40225fa2aceff4d6601dfc43bbb55fb01ee33b9cf8a
```

Afterwards `admin` user must be created with the `register` command and the newly generated token:

```bash
❯ https://example.local:3000/register?user=admin&token=5c8ccefbcb9f9a0ef8a8d40225fa2aceff4d6601dfc43bbb55fb01ee33b9cf8a
```

> ⚠️ The first step must be the `admin` user creation. Any other command will fail with an empty Redis database.

#### Action: `user-purge`

If somehow a user password has been lost, the `admin` can purge the user, allowing a new registration of the same username. For details, please see `register`.

Example:

```bash
❯ https://example.local:3000/admin?action=user-purge&name=<username_with_forgotten_pass>&token=<admin_token>
```

#### Action: `env-create`

Environments must exist for hosts to use host locking. This command can be used by the `admin` to create those dependency environments.

Example:

```bash
❯ https://example.local:3000/admin?action=env-create&name=<environment_name>&token=<admin_token>
```

#### Action: `env-unlock`

If some user locks an environment and it must be unlocked for any reason, the `admin` can do that. Environments and hosts normally can be unlocked by their owners.

Example:

```bash
❯ https://example.local:3000/admin?action=env-unlock&name=<environment_name>&token=<admin_token>
```

#### Action: `env-maintenance`

There are occasional maintenance timeframes. The `admin` can set it up. After the maintenance, it can be reverted by the `admin` with the `env-unlock` command.

Example:

```bash
❯ https://example.local:3000/admin?action=env-maintenance&name=<environment_name>&token=<admin_token>
```

#### Action: `env-terminate`

If an environment needs to be terminated for any reason, the `admin` can do that with the `env-terminate` action.

Example:

```bash
❯ https://example.local:3000/admin?action=env-terminate&name=<environment_name>&token=<admin_token>
```

#### Action: `host-unlock`

Host unlock is the same as `env-unlock` but for hosts.

Example:

```bash
❯ https://example.local:3000/admin?action=host-unlock&name=<environment_name>&token=<admin_token>
```

### Registering users

Except for the `stats` command, all other command needs a responsible user, which must be registered beforehand. Every user registers their user, no `admin` is needed for that.

```bash
❯ https://localhost:3000/register?user=<username>&token=<new_user_token>
```

### Locking hosts and environments

To own a host or an environment, it must be locked.

> ⚠️ Please be aware of the `lastday` parameter which describes the last day of the lock of the given host or env. RedisDB will release the lock automaticallyon the next day.

Examples:

```bash
❯ https://localhost:3000/lock?type=host&name=<hostname>&user=<username>&token=<user_token>&lastday=<expire_day>

❯ https://localhost:3000/lock?type=env&name=<envname>&user=<username>&token=<user_token>&lastday=<expore_day>
```

### Unlocking hosts and environments

Unlocking can be necessary sometimes before automatic unlocking happens, here is how to do that.

> ⚠️ Users cannot unlock envs or hosts owned by others. Only admin can do that.

Examples:

```bash
❯ https://localhost:3000/unlock?type=host&name=<hostname>&user=<username>&token=<user_token>

❯ https://localhost:3000/unlock?type=env&name=<envname>&user=<username>&token=<user_token>
```

### Status queries

To view the locking status for all environments and hosts, no special user validation is needed.

At this point, the HTML web page and JSON data response are supported.
Later YAML format will be added too.

#### Web HTML format (human readable)

```bash
❯ https://localhost:3000/status/web
```

#### JSON format

```bash
❯ https://localhost:3000/status/json
```

## Guarantees, responsibility

Please see the license:
https://github.com/drax2gma/nodelocker#Apache-2.0-1-ov-file
