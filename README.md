# NodeLocker

## Building the project

As this project was developed and tested on x86_64 linux, I'm providing a simple shell script to make binary building easier. Please see `./build_linux.sh` in the project's root directory.

## Running tests

> ⚠️ Please be aware that running the included ShellSpec tests **WILL DESTROY ALL YOUR LOCAL REDIS DATABASE**!
Use it with extra caution!

The test definitions are inside `tests/full_spec.sh`

You can read more on ShellSpec format here: https://shellspec.info/

Run the tests from the project directory with the following command:

$ `shellspec tests/`

## Running NodeLocker

After compiling with `./build_linux.sh` the output binary will be at `bin/release/nodelocker-linux`. When the local Redis database is ready, just run the compiled binary from there manually for testing.

```bash
./nodelocker-linux
```

Of course, the *lab-prod* behavior needs some extra setup, like a *systemd* module. I prefer *supervisord* for running such user-mode applications, your mileage may vary.

Just keep in mind, that if somehow the app fails, it won't restart itself, there is no watchdog feature implemented.

## Guarantees, responsibility
Please see the license:
https://github.com/drax2gma/nodelocker#Apache-2.0-1-ov-file