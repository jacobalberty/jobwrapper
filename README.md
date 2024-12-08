# JobWrapper

**JobWrapper** is a flexible wrapper for cron jobs that ensures script execution with configurable lock directories, error handling, and command output management. It prevents job conflicts using configurable locks, tracks job history, and integrates easily with custom scripts and configurations.

## Features

- **Flexible Locking**: Prevent job conflicts using configurable lock directories.
- **Error Handling**: Handles job execution errors and logs them for easier debugging.
- **History Tracking**: Tracks job execution history and outputs for later review.
- **Customizable Configurations**: Easily configure job settings via a configuration file.
- **Simple Integration**: Integrate easily with your cron job scripts.

## Installation

To install `jobwrapper`, clone the repository and build the project:

```bash
git clone https://github.com/jacobalberty/jobwrapper.git
cd jobwrapper
go build -o jobwrapper
```

Alternatively, you can install it using `go install`:

```bash
go install github.com/jacobalberty/jobwrapper@latest
```

## Usage

### Configuration

The configuration file (`jobwrapper.conf`) can be placed in `~/.jobwrapper` or any custom directory. Here's an example configuration:

```ini
lock_dir = "~/.jobwrapper"
timeout = 60
lock_filename = ".joblock"
history_lines = 5
```

### Running a Job

To run a job, execute `jobwrapper` with the appropriate arguments:

```bash
jobwrapper <group> <script> [args...]
```

- `<group>`: The group to which the job belongs (used for lock management).
- `<script>`: The script to be executed.
- `[args...]`: Optional arguments passed to the script.

Example:

```bash
jobwrapper backup /path/to/script.sh
```

### Cron Example

Here’s an example cron job using `jobwrapper`:

```bash
0 3 * * * /path/to/jobwrapper backup /path/to/backup_script.sh
```

This will run the `backup_script.sh` every day at 3:00 AM.

## License

`jobwrapper` is licensed under the MIT License. See [LICENSE](LICENSE) for more details.
